package main

import (
	"errors"
	"net/http"

	"alloy.dev/inertia-demo/auth"
	"alloy.dev/inertia-demo/crm"
	demoerrors "alloy.dev/inertia-demo/errors"
	"alloy.dev/inertia-demo/features"
	routegen "hara.sh/alloy/httpx/routing/navigator"
	"hara.sh/alloy/inertia"
	"hara.sh/alloy/inertia/protocol"
)

func initRoutes() *routegen.Registry {
	routes := routegen.New()

	routes.Add("login", "GET", "/login")

	routes.Add("logout", "POST", "/logout")

	crm.DefineRoutes(routes)

	features.DefineRoutes(routes)

	demoerrors.DefineRoutes(routes)

	return routes
}

func (rt *runtime) withDemoProps(authApp auth.App, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := authApp.CurrentUser(r)

		sidebarOpen := true

		if cookie, err := r.Cookie("sidebar_open"); err == nil {
			sidebarOpen = cookie.Value != "false"
		}

		ctx := r.Context()
		ctx = inertia.SetProps(ctx, protocol.Props{
			"sidebarOpen": sidebarOpen,
			"app": map[string]any{
				"name":        "Alloy Inertia Demo",
				"productLine": "Secondary Demo",
				"environment": "Demo",
			},
			"auth": map[string]any{
				"user": authApp.PublicUser(user),
			},
			"workspace": map[string]any{
				"name": "Alloy",
				"plan": "Inertia Protocol",
			},
			"routes": rt.routes.ManifestProps(),
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (rt *runtime) renderPage(w http.ResponseWriter, r *http.Request, component string, pageProps protocol.Props) {
	ctx := r.Context()

	if err := rt.inertia.Render(w, r.WithContext(ctx), component, pageProps); err != nil {
		switch {
		case errors.Is(err, protocol.ErrNotFound):
			http.Error(w, "page not found", http.StatusNotFound)
		default:
			http.Error(w, "demo internal error", http.StatusInternalServerError)
		}
	}
}
