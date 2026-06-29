package routing

import (
	"reflect"

	crouting "alloy.dev/backend/httpx/routing/contracts"
)

// RouteSignatureParameters extracts the parameter list of a route action's
// underlying callable, optionally filtered by a type predicate.
//
// In PHP this is implemented with ReflectionFunction / ReflectionMethod. In
// Go we use the [reflect] package: for a func value the parameter list is
// reflect.Type.In(i); for a "Handler@method" string we look the method up
// on the receiver's reflect.Type. Closures are unsupported because Go has no
// runtime parameter-name metadata for them — the parameter Type is recovered,
// but the parameter Name is always empty.
type RouteSignatureParameters struct{}

// SignatureParameter is a Go-friendly stand-in for ReflectionParameter.
type SignatureParameter struct {
	Index int
	Type  reflect.Type
	// Name is always "" for Go closures; populated for struct method receivers
	// only when the handler registry can map it (set by M5).
	Name string
}

// FromAction extracts signature parameters from a parsed [*Action].
//
//   - "subClass": when set, only parameters whose Type implements the named
//     interface (looked up via the handler registry) are returned. In M1
//     this is honored only for func handlers whose parameter types are
//     interface kinds.
//   - "backedEnum": when set, only parameters whose Type implements the
//     [BackedEnum] sentinel interface are returned.
//
// Returns parameters in declaration order.

// BackedEnum is the sentinel interface that user-defined backed enums must
// implement to participate in signature filtering.
type BackedEnum = crouting.BackedEnum

func (RouteSignatureParameters) FromAction(action *Action, conditions map[string]any) []SignatureParameter {
	if action == nil || action.Uses == nil {
		return nil
	}

	rv := reflect.ValueOf(action.Uses)

	if rv.Kind() != reflect.Func {
		return nil
	}

	rt := rv.Type()
	out := make([]SignatureParameter, 0, rt.NumIn())

	for i := 0; i < rt.NumIn(); i++ {
		out = append(out, SignatureParameter{Index: i, Type: rt.In(i)})
	}

	return filterParameters(out, conditions)
}

func filterParameters(in []SignatureParameter, conditions map[string]any) []SignatureParameter {
	if len(conditions) == 0 {
		return in
	}

	if _, ok := conditions["backedEnum"]; ok {
		backedEnumType := reflect.TypeOf((*BackedEnum)(nil)).Elem()
		out := in[:0]

		for _, p := range in {
			if p.Type.Implements(backedEnumType) {
				out = append(out, p)
			}
		}

		return out
	}

	if sub, ok := conditions["subClass"]; ok {
		if subType, ok := sub.(reflect.Type); ok {
			out := in[:0]

			for _, p := range in {
				if p.Type.Implements(subType) || (p.Type.Kind() == reflect.Ptr && p.Type.Elem().Implements(subType)) {
					out = append(out, p)
				}
			}

			return out
		}
	}

	return in
}
