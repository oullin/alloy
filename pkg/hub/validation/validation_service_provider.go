package validation

import "hara.sh/alloy/container"

// ValidationServiceProvider registers the validator factory into the container.
// Ref: @alloy/code-0390
type ValidationServiceProvider struct {
	app *container.App
}

// NewValidationServiceProvider constructs the provider.
func NewValidationServiceProvider(app *container.App) *ValidationServiceProvider {
	return &ValidationServiceProvider{app: app}
}

// Register binds the validator factory as a singleton under "validator".
func (p *ValidationServiceProvider) Register() {
	p.app.Singleton("validator", func(_ *container.App) (any, error) {
		return NewFactory(), nil
	})
}

// Provides returns the abstract keys registered by this provider.
func (p *ValidationServiceProvider) Provides() []string {
	return []string{"validator"}
}
