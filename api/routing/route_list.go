package routing

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
)

// RouteListEntry is one displayable route:list row.
type RouteListEntry struct {
	Domain     string   `json:"domain"`
	Method     string   `json:"method"`
	URI        string   `json:"uri"`
	Name       string   `json:"name"`
	Action     string   `json:"action"`
	Middleware []string `json:"middleware"`
	Path       string   `json:"path"`
	Vendor     bool     `json:"vendor"`
}

// RouteListOptions controls route-list filtering, ordering, and output shape.
type RouteListOptions struct {
	Method     string
	Action     string
	Name       string
	Domain     string
	Middleware string
	Path       string
	ExceptPath string

	Sort    string
	Reverse bool
	Columns []string

	ExceptVendor bool
	OnlyVendor   bool

	MiddlewareMode MiddlewareGatherMode
}

// ListRoutes returns Laravel-style route:list entries for the router.
func (r *Router) ListRoutes(opts RouteListOptions) []RouteListEntry {
	return ListRoutes(r, r.GetRoutes(), opts)
}

// RouteListJSON returns selected route-list columns as JSON.
func (r *Router) RouteListJSON(opts RouteListOptions) ([]byte, error) {
	return json.Marshal(r.RouteListRows(opts))
}

// RouteListRows returns selected route-list columns as map rows.
func (r *Router) RouteListRows(opts RouteListOptions) []map[string]any {
	entries := r.ListRoutes(opts)
	rows := make([]map[string]any, 0, len(entries))

	for _, entry := range entries {
		rows = append(rows, entry.RouteListColumns(opts.Columns))
	}

	return rows
}

// ListRoutes returns Laravel-style route:list entries for a collection.
func ListRoutes(router *Router, collection RouteCollectionInterface, opts RouteListOptions) []RouteListEntry {
	routes := collection.GetRoutes()
	entries := make([]RouteListEntry, 0, len(routes))

	for _, route := range routes {
		entry := routeListEntry(router, route, opts.MiddlewareMode)

		if routeListFiltered(entry, opts) {
			continue
		}

		entries = append(entries, entry)
	}

	sortRouteList(entries, opts.Sort)

	if opts.Reverse {
		reverseRouteList(entries)
	}

	return entries
}

// RouteListColumns returns a selected-column map for the entry. When columns
// is empty, every public route-list field is included.
func (e RouteListEntry) RouteListColumns(columns []string) map[string]any {
	if len(columns) == 0 {
		columns = []string{"domain", "method", "uri", "name", "action", "middleware", "path", "vendor"}
	}

	row := make(map[string]any, len(columns))

	for _, column := range normalizeRouteListColumns(columns) {
		switch column {
		case "domain":
			row[column] = e.Domain
		case "method":
			row[column] = e.Method
		case "uri":
			row[column] = e.URI
		case "name":
			row[column] = e.Name
		case "action":
			row[column] = e.Action
		case "middleware":
			row[column] = append([]string(nil), e.Middleware...)
		case "path":
			row[column] = e.Path
		case "vendor":
			row[column] = e.Vendor
		}
	}

	return row
}

func routeListEntry(router *Router, route *Route, mode MiddlewareGatherMode) RouteListEntry {
	if route == nil {
		return RouteListEntry{}
	}

	path, vendor := routeActionPath(route)
	middleware := []string{}

	if router != nil {
		middleware = routeListMiddleware(router.GatherRouteMiddleware(route, mode))
	}

	return RouteListEntry{
		Domain:     route.GetDomain(),
		Method:     routeListMethod(route),
		URI:        routeListURI(route),
		Name:       route.GetName(),
		Action:     strings.TrimLeft(route.GetActionName(), `\`),
		Middleware: middleware,
		Path:       path,
		Vendor:     vendor,
	}
}

func routeListMethod(route *Route) string {
	method := strings.Join(route.Methods(), "|")

	if method == strings.Join(HTTPVerbs, "|") {
		return "ANY"
	}

	return method
}

func routeListURI(route *Route) string {
	uri := route.Uri

	for parameter, field := range route.BindingFields() {
		uri = strings.ReplaceAll(uri, "{"+parameter+"}", "{"+parameter+":"+field+"}")
		uri = strings.ReplaceAll(uri, "{"+parameter+"?}", "{"+parameter+":"+field+"?}")
	}

	return uri
}

func routeListMiddleware(middleware []any) []string {
	out := make([]string, 0, len(middleware))

	for _, item := range middleware {
		if item == nil {
			continue
		}

		if _, ok := item.(func()); ok {
			out = append(out, "Closure")

			continue
		}

		if name, ok := item.(string); ok {
			out = append(out, name)

			continue
		}

		out = append(out, reflect.TypeOf(item).String())
	}

	return out
}

func routeActionPath(route *Route) (string, bool) {
	uses := route.ActionMap["uses"]
	value := reflect.ValueOf(uses)

	if !value.IsValid() || value.Kind() != reflect.Func {
		return "", false
	}

	fn := runtime.FuncForPC(value.Pointer())

	if fn == nil {
		return "", false
	}

	file, line := fn.FileLine(value.Pointer())
	path := filepath.ToSlash(file)

	if line > 0 {
		path += ":" + intString(line)
	}

	return path, isVendorRoutePath(path)
}

func isVendorRoutePath(path string) bool {
	path = filepath.ToSlash(path)

	return strings.Contains(path, "/vendor/") || strings.Contains(path, "/pkg/mod/")
}

func routeListFiltered(entry RouteListEntry, opts RouteListOptions) bool {
	middleware := strings.Join(entry.Middleware, "\n")

	if opts.Name != "" && !strings.Contains(entry.Name, opts.Name) {
		return true
	}

	if opts.Action != "" && !strings.Contains(entry.Action, opts.Action) {
		return true
	}

	if opts.Path != "" && !strings.Contains(entry.URI, opts.Path) {
		return true
	}

	if opts.Method != "" && !routeListMethodMatches(entry.Method, opts.Method) {
		return true
	}

	if opts.Domain != "" && !strings.Contains(entry.Domain, opts.Domain) {
		return true
	}

	if opts.Middleware != "" && !strings.Contains(middleware, opts.Middleware) {
		return true
	}

	if opts.ExceptVendor && entry.Vendor {
		return true
	}

	if opts.OnlyVendor && !entry.Vendor {
		return true
	}

	for _, path := range strings.Split(opts.ExceptPath, ",") {
		path = strings.TrimSpace(path)

		if path != "" && strings.Contains(entry.URI, path) {
			return true
		}
	}

	return false
}

func sortRouteList(entries []RouteListEntry, sortBy string) {
	if strings.TrimSpace(sortBy) == "" {
		sortBy = "uri"
	}

	if sortBy == "definition" {
		return
	}

	keys := normalizeRouteListColumns(strings.Split(sortBy, ","))

	sort.SliceStable(entries, func(i, j int) bool {
		for _, key := range keys {
			left := routeListSortValue(entries[i], key)
			right := routeListSortValue(entries[j], key)

			if left == right {
				continue
			}

			return left < right
		}

		return false
	})
}

func routeListSortValue(entry RouteListEntry, key string) string {
	switch key {
	case "domain":
		return entry.Domain
	case "method":
		return entry.Method
	case "uri":
		return entry.URI
	case "name":
		return entry.Name
	case "action":
		return entry.Action
	case "middleware":
		return strings.Join(entry.Middleware, "\n")
	case "path":
		return entry.Path
	case "vendor":
		if entry.Vendor {
			return "1"
		}

		return "0"
	default:
		return entry.URI
	}
}

func routeListMethodMatches(routeMethod, filter string) bool {
	filter = strings.ToUpper(strings.TrimSpace(filter))

	if filter == "" {
		return true
	}

	filters := strings.FieldsFunc(filter, func(r rune) bool {
		return r == ',' || r == '|'
	})

	if routeMethod == "ANY" {
		for _, method := range filters {
			if strings.TrimSpace(method) == "ANY" {
				return true
			}
		}

		return false
	}

	routeMethods := strings.Split(routeMethod, "|")

	for _, routeMethod := range routeMethods {
		routeMethod = strings.TrimSpace(routeMethod)

		for _, method := range filters {
			if routeMethod == strings.TrimSpace(method) {
				return true
			}
		}
	}

	return false
}

func reverseRouteList(entries []RouteListEntry) {
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
}

func normalizeRouteListColumns(columns []string) []string {
	out := make([]string, 0, len(columns))

	for _, column := range columns {
		for _, part := range strings.Split(column, ",") {
			part = strings.ToLower(strings.TrimSpace(part))

			if part != "" {
				out = append(out, part)
			}
		}
	}

	return out
}

func intString(value int) string {
	if value == 0 {
		return "0"
	}

	negative := value < 0

	if negative {
		value = -value
	}

	var buf [20]byte
	pos := len(buf)

	for value > 0 {
		pos--
		buf[pos] = byte('0' + value%10)
		value /= 10
	}

	if negative {
		pos--
		buf[pos] = '-'
	}

	return string(buf[pos:])
}
