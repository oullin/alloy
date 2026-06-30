package main

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	routegen "alloy.dev/go/httpx/routing/navigator"
	"alloy.dev/go/inertia"
	"alloy.dev/go/inertia/flash"
	"alloy.dev/go/inertia/middleware"
	corei18n "alloy.dev/go/seo/i18n"
	"alloy.dev/inertia-demo/internal/database"
	"alloy.dev/inertia-demo/internal/seed"
)

//go:embed resources/views/app.html
var rootTemplateFS embed.FS

type runtime struct {
	db         *sql.DB
	cryptoKey  []byte
	inertia    *inertia.Inertia
	localeCfg  *corei18n.I18nConfig
	flashStore *flash.CookieStore
	routes     *routegen.Registry
}

func main() {
	tmpl, err := rootTemplateFS.ReadFile("resources/views/app.html")

	if err != nil {
		log.Fatal("failed to read template:", err)
	}

	version := "dev"

	if v := os.Getenv("APP_VERSION"); strings.TrimSpace(v) != "" {
		version = v
	}

	seoPath, err := resolveResourcePath("seo.yml")

	if err != nil {
		log.Fatal(err)
	}

	i, err := inertia.New(string(tmpl),
		inertia.WithVersion(version),
		inertia.WithHeadFromFile(seoPath),
	)

	if err != nil {
		log.Fatal(err)
	}

	localeCfg, err := LoadI18n(mustResolveResourcePath("i18n.yml"))

	if err != nil {
		log.Fatal(err)
	}

	// The demo keeps canonical, non-prefixed routes in the frontend while
	// still consuming locale-driven head defaults from config.
	localeCfg.URLPrefix = false

	cryptoCfg, err := LoadCrypto(mustResolveResourcePath("crypto.yml"))

	if err != nil {
		log.Fatal(err)
	}

	cryptoKey, err := cryptoCfg.DecodedKey()

	if err != nil {
		log.Fatal(err)
	}

	csrfCfg, err := middleware.LoadCSRF(mustResolveResourcePath("csrf.yml"))

	if err != nil {
		log.Fatal(err)
	}

	csrfMiddleware := middleware.CSRF(csrfCfg, cryptoKey)

	dbPath, err := resolveDatabasePath()

	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Open(dbPath)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	if err := seed.Run(db); err != nil {
		log.Fatal("failed to seed database:", err)
	}

	rt := &runtime{
		db:         db,
		cryptoKey:  cryptoKey,
		inertia:    i,
		localeCfg:  localeCfg,
		flashStore: flash.NewCookieStore(flash.WithCookieName("beacon_flash"), flash.WithKey(cryptoKey)),
		routes:     initRoutes(),
	}

	distPath, err := resolveDistPath()

	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.Handle(
		"/assets/",
		http.StripPrefix("/assets/", http.FileServer(http.Dir(distPath))),
	)

	appMux := http.NewServeMux()

	authApp, err := rt.newAuth()

	if err != nil {
		log.Fatalf("auth: %v", err)
	}

	authApp.RegisterRoutes(appMux)

	if err := rt.registerCRMRoutes(appMux, authApp); err != nil {
		log.Fatalf("crm routes: %v", err)
	}

	if err := rt.registerFeatureRoutes(appMux, authApp); err != nil {
		log.Fatalf("feature routes: %v", err)
	}

	if err := rt.registerErrorRoutes(appMux, authApp); err != nil {
		log.Fatalf("error routes: %v", err)
	}

	appMux.Handle("GET /{$}", http.RedirectHandler("/dashboard", http.StatusFound))

	mux.Handle("/", rt.dashboardAppHandler(authApp.WithCurrentUser(rt.withDemoProps(authApp, appMux)), csrfMiddleware))

	addr := ":8080"

	if port := os.Getenv("PORT"); strings.TrimSpace(port) != "" {
		addr = ":" + port
	}

	if url := os.Getenv("PORTLESS_URL"); strings.TrimSpace(url) != "" {
		fmt.Printf("Server running at %s\n", url)
	} else {
		fmt.Printf("Server running at http://localhost%s\n", addr)
	}

	log.Fatal(http.ListenAndServe(addr, mux))
}

func resolveDistPath() (string, error) {
	if path := strings.TrimSpace(os.Getenv("ALLOY_INERTIA_DIST_PATH")); path != "" {
		return filepath.Clean(path), nil
	}

	candidates := []string{
		filepath.Join("web", "storage", "inertia-demo", "dist", "app"),
		filepath.Join("..", "..", "storage", "inertia-demo", "dist", "app"),
		"../app/dist",
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return filepath.Clean(candidate), nil
		}
	}

	return "", fmt.Errorf("failed to locate demo app dist directory")
}

func resolveDatabasePath() (string, error) {
	if path := strings.TrimSpace(os.Getenv("ALLOY_INERTIA_DB_PATH")); path != "" {
		return filepath.Clean(path), nil
	}

	candidates := []string{
		filepath.Join("web", "storage", "inertia-demo", "beacon.db"),
		filepath.Join("..", "..", "storage", "inertia-demo", "beacon.db"),
	}

	for _, candidate := range candidates {
		dir := filepath.Dir(candidate)
		parent := filepath.Dir(dir)

		if info, err := os.Stat(parent); err == nil && info.IsDir() {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return "", err
			}

			return filepath.Clean(candidate), nil
		}
	}

	return "", fmt.Errorf("failed to resolve demo database path")
}

func resolveResourcePath(name string) (string, error) {
	candidates := []string{
		filepath.Join("cmd", "resources", name),
		filepath.Join("resources", name),
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return filepath.Clean(candidate), nil
		}
	}

	return "", fmt.Errorf("failed to locate demo resource %q", name)
}

func mustResolveResourcePath(name string) string {
	path, err := resolveResourcePath(name)

	if err != nil {
		log.Fatal(err)
	}

	return path
}

func (rt *runtime) dashboardAppHandler(base http.Handler, csrfMiddleware func(http.Handler) http.Handler) http.Handler {
	handler := base

	handler = flash.Middleware(rt.flashStore)(handler)

	if rt.localeCfg != nil {
		handler = corei18n.Middleware(rt.localeCfg, handler)
	}

	if csrfMiddleware != nil {
		handler = csrfMiddleware(handler)
	}

	handler = middleware.HTTPPreview()(handler)

	return rt.inertia.Middleware(handler)
}
