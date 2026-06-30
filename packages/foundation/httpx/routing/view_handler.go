package routing

import crouting "alloy.dev/foundation/httpx/routing/contracts"

// ViewFactory is the minimum surface ViewHandler needs from the bedrock
// view layer. The service provider in M11 wires in a real implementation.
type ViewFactory = crouting.ViewFactory

// ViewHandler is the invokable handler used by [Router.View]. It
// renders the named view with the supplied data, merging in any extra route
// parameters as additional data.
type ViewHandler struct {
	Handler
	view ViewFactory
}

// NewViewHandler wires the handler to a view factory.
func NewViewHandler(view ViewFactory) *ViewHandler {
	return &ViewHandler{view: view}
}

// Invoke dispatches the view with the route parameters merged into data.
func (c *ViewHandler) Invoke(route *Route, viewName string, data map[string]any) any {
	if data == nil {
		data = map[string]any{}
	}

	for k, v := range route.Parameters {
		if k == "view" || k == "data" || k == "status" || k == "headers" {
			continue
		}

		data[k] = v
	}

	if c.view == nil {
		return nil
	}

	return c.view.Make(viewName, data)
}
