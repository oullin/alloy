package main

import (
	"net/http"

	"alloy.dev/apps/inertia/api/auth"
	apierrors "alloy.dev/apps/inertia/api/errors"
)

func (rt *runtime) registerErrorRoutes(mux *http.ServeMux, authApp auth.App) error {
	return apierrors.RegisterRoutes(rt.routes, mux, apierrors.Container{
		RequireAuth: authApp.RequireAuth,
		Render:      rt.renderPage,
	})
}
