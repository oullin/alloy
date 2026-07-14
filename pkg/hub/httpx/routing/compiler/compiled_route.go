package compiler

import ccompiler "github.com/oullin/alloy/pkg/hub/httpx/routing/contracts/compiler"

// Token is one element of a compiled route pattern.
type Token = ccompiler.Token

// CompiledRoute is the result of compiling a route URI and optional host.
type CompiledRoute = ccompiler.CompiledRoute

// NewCompiledRoute constructs a CompiledRoute.
var NewCompiledRoute = ccompiler.NewCompiledRoute
