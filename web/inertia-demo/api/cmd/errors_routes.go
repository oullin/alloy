package main

import (
	"net/http"

	"alloy.dev/inertia-demo/auth"
	apierrors "alloy.dev/inertia-demo/errors"
)

func (rt *runtime) registerErrorRoutes(mux *http.ServeMux, authApp auth.App) error {
	return apierrors.RegisterRoutes(rt.routes, mux, apierrors.Container{
		RequireAuth: authApp.RequireAuth,
		Render:      rt.renderPage,
	})
}
