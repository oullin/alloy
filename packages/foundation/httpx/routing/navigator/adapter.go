package navigator

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/oullin/alloy/packages/foundation/httpx/routing"
)

// AdapterOptions configures how Alloy routes are converted to RouteInfo.
type AdapterOptions struct {
	// AppURL is the base application URL (e.g. "http://localhost:8000/v2").
	// When set, its path component is prepended to all generated URIs,
	// and its host/scheme override routes that do not have their own domain.
	AppURL string

	// ForcedScheme overrides the scheme for all routes ("http://", "https://").
	ForcedScheme string

	// ForcedRoot overrides the host, scheme, and base path for routes that do
	// not declare their own domain.
	ForcedRoot string

	// Defaults holds URL default values, equivalent to URL::defaults() values
	// discovered from route middleware.
	Defaults map[string]string
}

// FromRouteCollection converts an Alloy RouteCollectionInterface into a slice
// of RouteInfo values suitable for passing to Generate.
func FromRouteCollection(collection routing.RouteCollectionInterface, opts AdapterOptions) []*RouteInfo {
	routes := collection.GetRoutes()
	out := make([]*RouteInfo, 0, len(routes))

	basePath, forcedScheme := parseAppURL(opts.AppURL)
	rootDomain, rootScheme, rootBasePath := parseForcedRoot(opts.ForcedRoot)

	if rootBasePath != "" {
		basePath = rootBasePath
	}

	if rootScheme != "" {
		forcedScheme = rootScheme
	}

	if opts.ForcedScheme != "" {
		forcedScheme = normalizeScheme(opts.ForcedScheme)
	}

	for _, r := range routes {
		info := adaptRoute(r, basePath, forcedScheme, rootDomain, opts.Defaults)
		out = append(out, info)
	}

	return out
}

// GenerateFromRouteCollection adapts a live route collection and writes
// TypeScript route exposure files.
func GenerateFromRouteCollection(collection routing.RouteCollectionInterface, generate Options, adapter AdapterOptions) error {
	return Generate(FromRouteCollection(collection, adapter), generate)
}

// GenerateFromRouter adapts a live router and writes TypeScript route exposure
// files from its registered route collection.
func GenerateFromRouter(router *routing.Router, generate Options, adapter AdapterOptions) error {
	if router == nil {
		return fmt.Errorf("expose: nil router")
	}

	return GenerateFromRouteCollection(router.GetRoutes(), generate, adapter)
}

// adaptRoute converts a single routing.Route to a RouteInfo.
func adaptRoute(r *routing.Route, basePath, forcedScheme, rootDomain string, urlDefaults map[string]string) *RouteInfo {
	// Lower-case HTTP methods.
	methods := make([]string, len(r.HTTPMethods))

	for i, m := range r.HTTPMethods {
		methods[i] = strings.ToLower(m)
	}

	// Handler string ("Class@Method" or "Closure").
	handler := r.GetActionName()
	actionMethod := r.GetActionMethod()

	// Invokable detection: the Go routing package registers invokable
	// handlers with method "Invoke" (capital I) appended by ParseAction.
	isInvokable := actionMethod == "Invoke" || actionMethod == "__invoke"

	// Route name — strip generated:: prefix and trailing dots.
	name := cleanRouteName(r.GetName())

	// Parameters: walk ParameterNames and check URI for optional marker.
	paramNames := r.ParameterNames()
	bindingFields := r.BindingFields()
	defaults := r.DefaultValues
	defaultStrings := mergeDefaultStrings(urlDefaults, defaults)
	params := make([]Param, 0, len(paramNames))

	for _, pName := range paramNames {
		optional := isOptionalParam(r.Uri, pName) || hasDefault(defaults, pName) || hasStringDefault(urlDefaults, pName)
		key := bindingFields[pName]
		defaultVal := defaultStrings[pName]

		params = append(params, Param{
			Name:     pName,
			Optional: optional,
			Key:      key,
			Default:  defaultVal,
		})
	}

	// Domain handling.
	domain := r.GetDomain()
	scheme := ""

	if domain == "" {
		domain = rootDomain
	}

	if domain != "" {
		scheme = "//"
	}

	if forcedScheme != "" {
		scheme = forcedScheme
	}

	if r.HttpOnly() {
		scheme = "http://"
	}

	if r.HttpsOnly() {
		scheme = "https://"
	}

	return &RouteInfo{
		URI:         r.Uri,
		Methods:     methods,
		Name:        name,
		Handler:     handler,
		IsInvokable: isInvokable,
		Params:      params,
		Domain:      domain,
		Scheme:      scheme,
		BasePath:    basePath,
		Defaults:    defaultStrings,
	}
}

// isOptionalParam reports whether the named parameter appears as {name?}
// (with trailing question mark) in the URI string.
func isOptionalParam(uri, name string) bool {
	return strings.Contains(uri, "{"+name+"?}")
}

// hasDefault reports whether name has an entry in the defaults map.
func hasDefault(defaults map[string]any, name string) bool {
	if defaults == nil {
		return false
	}

	_, ok := defaults[name]

	return ok
}

func hasStringDefault(defaults map[string]string, name string) bool {
	if defaults == nil {
		return false
	}

	_, ok := defaults[name]

	return ok
}

func mergeDefaultStrings(urlDefaults map[string]string, routeDefaults map[string]any) map[string]string {
	if len(urlDefaults) == 0 && len(routeDefaults) == 0 {
		return nil
	}

	out := make(map[string]string, len(urlDefaults)+len(routeDefaults))

	for key, value := range urlDefaults {
		out[key] = value
	}

	for key := range routeDefaults {
		out[key] = defaultString(routeDefaults, key)
	}

	return out
}

// defaultString returns the string representation of a default value,
// or empty string when none exists.
func defaultString(defaults map[string]any, name string) string {
	if defaults == nil {
		return ""
	}

	v, ok := defaults[name]

	if !ok {
		return ""
	}

	switch s := v.(type) {
	case string:
		return s
	case int:
		return strconv.Itoa(s)
	default:
		return ""
	}
}

// parseAppURL extracts the base path and scheme from an APP_URL string.
// "http://localhost:8000/v2" → basePath="/v2", scheme=""
// "https://example.com"     → basePath="", scheme=""
func parseAppURL(appURL string) (basePath, scheme string) {
	if appURL == "" {
		return "", ""
	}

	if strings.HasPrefix(appURL, "//") {
		appURL = "http:" + appURL
	}

	u, err := url.Parse(appURL)

	if err != nil {
		return "", ""
	}

	path := strings.TrimRight(u.Path, "/")

	if path == "/" || path == "" {
		return "", ""
	}

	return path, ""
}

func parseForcedRoot(root string) (domain, scheme, basePath string) {
	if root == "" {
		return "", "", ""
	}

	if strings.HasPrefix(root, "//") {
		root = "http:" + root
	}

	u, err := url.Parse(root)

	if err != nil {
		return "", "", ""
	}

	if u.Host != "" {
		domain = u.Host
	}

	if u.Scheme != "" {
		scheme = normalizeScheme(u.Scheme)
	}

	path := strings.TrimRight(u.Path, "/")

	if path != "/" {
		basePath = path
	}

	return domain, scheme, basePath
}

func normalizeScheme(scheme string) string {
	if scheme == "" {
		return ""
	}

	if strings.HasSuffix(scheme, "://") || scheme == "//" {
		return scheme
	}

	return strings.TrimSuffix(scheme, ":") + "://"
}
