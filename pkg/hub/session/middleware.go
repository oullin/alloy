package session

import (
	"context"
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
	// Secure sets the Secure flag on the session cookie. It is a tri-state:
	// nil (the default) means secure-by-default (true); an explicit false is
	// honored so local HTTP dev can opt out. Use BoolPtr to set it.
	Secure *bool
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

// BoolPtr returns a pointer to v, for setting the tri-state Secure config.
func BoolPtr(v bool) *bool { return &v }

func defaultConfig() StartSessionConfig {
	return StartSessionConfig{
		CookieName:    "session",
		Lifetime:      2 * time.Hour,
		Secure:        BoolPtr(true),
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

	// Tri-state: nil keeps the secure-by-default; an explicit value (true or
	// false) is honored so local HTTP dev can opt out with BoolPtr(false).
	if cfg.Secure != nil {
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
		Secure:   w.cfg.Secure == nil || *w.cfg.Secure,
		SameSite: w.cfg.SameSite,
	})
}

// StartSession is HTTP middleware that manages the full session lifecycle for
// each request: read → start → handle → save → write cookie.
//
// Background GC is scheduled against a background context. Use
// StartSessionWithContext to bind GC to a server lifecycle so it stops cleanly
// on shutdown.
func StartSession(handler Handler, cfg StartSessionConfig) func(http.Handler) http.Handler {
	return StartSessionWithContext(context.Background(), handler, cfg)
}

// StartSessionWithContext is StartSession with an explicit lifecycle context
// that bounds background session GC: when ctx is cancelled (e.g. on server
// shutdown) no new sweep starts and any in-flight sweep is cancelled through
// the context handed to the handler.
func StartSessionWithContext(ctx context.Context, handler Handler, cfg StartSessionConfig) func(http.Handler) http.Handler {
	cfg = mergeConfig(cfg)
	gc := newGCScheduler(ctx, handler, cfg.GCMaxLifetime)

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

			// Probabilistic GC, scheduled off the request path so the
			// directory walk / bulk DELETE never lands in this request's
			// latency budget. The scheduler is single-flight and panic-safe.
			if cfg.GCProbability > 0 && rand.IntN(100) < cfg.GCProbability {
				gc.trigger()
			}
		})
	}
}
