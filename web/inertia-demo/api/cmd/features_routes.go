package main

import (
	"net/http"

	"alloy.dev/inertia-demo/auth"
	"alloy.dev/inertia-demo/features"
)

func (rt *runtime) registerFeatureRoutes(mux *http.ServeMux, authApp auth.App) error {
	redirectFn := func(w http.ResponseWriter, r *http.Request, url string) {
		rt.inertia.Redirect(w, r, url)
	}

	locationFn := func(w http.ResponseWriter, r *http.Request, url string) {
		rt.inertia.Location(w, r, url)
	}

	return features.RegisterRoutes(rt.routes, mux, features.Container{
		DB:          rt.db,
		RequireAuth: authApp.RequireAuth,
		Render:      rt.renderPage,
		Redirect:    redirectFn,

		Location: locationFn,

		RouteURL: rt.routes.URL,
		SetFlash: rt.flashStore.Set,
	})
}
