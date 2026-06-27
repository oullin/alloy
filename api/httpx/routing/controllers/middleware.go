package controllers

import ccontrollers "github.com/oullin/alloy/api/httpx/routing/contracts/controllers"

// Middleware describes controller middleware and method filters.
type Middleware = ccontrollers.Middleware

// NewMiddleware constructs a Middleware value with no method filter.
var NewMiddleware = ccontrollers.NewMiddleware
