package session

import (
	"math/rand/v2"
	"net/http"
	"time"
)

// StartSessionConfig configures the StartSession middleware.
type StartSessionConfig struct {
	// CookieName is the session cookie name (default: "session").
	CookieName string
	// Lifetime is the session cookie max-age (default: 2 hours).
	Lifetime time.Duration
	// Secure sets the Secure flag on the session cookie.
	Secure bool
	// SameSite is the SameSite attribute for the session cookie.
	SameSite http.SameSite
	// GCProbability controls the chance (0–100) of running GC on a request.
	GCProbability int
	// GCMaxLifetime is the max session age in seconds for GC.
	GCMaxLifetime int
	// ActivityRefresh is the sliding-expiry interval: on a read-only request,
	// the session's backend record is refreshed at most once per this interval
	// (rather than once per request) so an active user's session does not
	// expire while unchanged. A negative value disables the refresh entirely.
	// When zero, it defaults to half the Lifetime.
	ActivityRefresh time.Duration
}

// sessionResponseWriter intercepts WriteHeader / Write to flush the session
// cookie before the response headers are committed.
type sessionResponseWriter struct {
	http.ResponseWriter
	store   *Store
	cfg     StartSessionConfig
	flushed bool
}

func defaultConfig() StartSessionConfig {
	return StartSessionConfig{
		CookieName:    "session",
		Lifetime:      2 * time.Hour,
		SameSite:      http.SameSiteLaxMode,
		GCProbability: 2,
		GCMaxLifetime: 7200,
	}
}

func mergeConfig(cfg StartSessionConfig) StartSessionConfig {
	defaults := defaultConfig()

	if cfg.CookieName != "" {
		defaults.CookieName = cfg.CookieName
	}

	if cfg.Lifetime != 0 {
		defaults.Lifetime = cfg.Lifetime
	}

	if cfg.Secure {
		defaults.Secure = cfg.Secure
	}

	if cfg.SameSite != 0 {
		defaults.SameSite = cfg.SameSite
	}

	if cfg.GCProbability != 0 {
		defaults.GCProbability = cfg.GCProbability
	}

	if cfg.GCMaxLifetime != 0 {
		defaults.GCMaxLifetime = cfg.GCMaxLifetime
	}

	// Sliding-expiry refresh: zero means "derive from Lifetime" so an active
	// session is kept alive without a per-request write; a negative value is an
	// explicit opt-out and is preserved as-is (TouchActivity treats it as off).
	switch {
	case cfg.ActivityRefresh < 0:
		defaults.ActivityRefresh = cfg.ActivityRefresh
	case cfg.ActivityRefresh > 0:
		defaults.ActivityRefresh = cfg.ActivityRefresh
	default:
		defaults.ActivityRefresh = defaults.Lifetime / 2
	}

	return defaults
}

func (w *sessionResponseWriter) WriteHeader(code int) {
	w.flush()
	w.ResponseWriter.WriteHeader(code)
}

func (w *sessionResponseWriter) Write(b []byte) (int, error) {
	w.flush()

	return w.ResponseWriter.Write(b)
}

func (w *sessionResponseWriter) flush() {
	if w.flushed {
		return
	}

	w.flushed = true
	maxAge := int(w.cfg.Lifetime.Seconds())
	http.SetCookie(w.ResponseWriter, &http.Cookie{
		Name:     w.cfg.CookieName,
		Value:    w.store.GetID(),
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   w.cfg.Secure,
		SameSite: w.cfg.SameSite,
	})
}

// StartSession is HTTP middleware that manages the full session lifecycle for
// each request: read → start → handle → save → write cookie.
func StartSession(handler Handler, cfg StartSessionConfig) func(http.Handler) http.Handler {
	cfg = mergeConfig(cfg)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Resolve session ID from incoming cookie.
			var id string

			if c, err := r.Cookie(cfg.CookieName); err == nil {
				id = c.Value
			}

			var store *Store

			if id != "" {
				store = NewWithID(cfg.CookieName, handler, id)
			} else {
				store = New(cfg.CookieName, handler)
			}

			// Inform existence-aware handlers whether the session already exists.
			if ea, ok := handler.(ExistenceAware); ok {
				ea.SetExists(id != "")
			}

			if err := store.Start(r.Context()); err != nil {
				http.Error(w, "session start failed", http.StatusInternalServerError)

				return
			}

			// Expose the Store to downstream handlers via the request context.
			r = r.WithContext(NewContext(r.Context(), store))

			sw := &sessionResponseWriter{ResponseWriter: w, store: store, cfg: cfg}
			next.ServeHTTP(sw, r)

			// Sliding expiry: refresh the activity marker at most once per
			// ActivityRefresh interval so an unchanged but active session is
			// still persisted periodically. Save itself skips the write when
			// the session is clean.
			store.TouchActivity(time.Now(), cfg.ActivityRefresh)

			if err := store.Save(r.Context()); err != nil {
				// Best-effort; session save errors should not abort the response.
				_ = err
			}

			// If the handler never wrote, flush the cookie now.
			sw.flush()

			// Probabilistic GC.
			if cfg.GCProbability > 0 && rand.IntN(100) < cfg.GCProbability {
				_ = handler.GC(r.Context(), cfg.GCMaxLifetime)
			}
		})
	}
}
