package cookie

import "github.com/oullin/alloy/container"

// CookieServiceProvider registers the cookie jar into the container.
// Ref: @bedrock/code-0200
type CookieServiceProvider struct {
	app      *container.Container
	defaults Options
}

// NewCookieServiceProvider constructs the provider with the given default cookie options.
// Pass DefaultOptions() to use sensible production defaults.
func NewCookieServiceProvider(app *container.Container, defaults Options) *CookieServiceProvider {
	return &CookieServiceProvider{app: app, defaults: defaults}
}

// Register binds a fresh cookie jar factory under "cookie".
func (p *CookieServiceProvider) Register() {
	p.app.Bind("cookie", func(_ *container.Container) (any, error) {
		return NewJar(p.defaults), nil
	}, false)
}

// Provides returns the abstract keys registered by this provider.
func (p *CookieServiceProvider) Provides() []string {
	return []string{"cookie"}
}
