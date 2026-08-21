package config

import (
	"fmt"
	"strings"
)

// KeySet is the set of dotted keys a configuration file actually wrote,
// including list indices ("providers.0.enabled").
//
// It exists because for a boolean that defaults to true, the Go zero value and
// an explicit "false" are indistinguishable after unmarshalling. Merging on the
// zero value alone would make `require-pushed-commits: false` a no-op, which is
// the worst possible direction for that particular flag to fail in. Viper's own
// AllKeys does not descend into lists, so the set is built by walking the
// settings map directly.
type KeySet map[string]struct{}

// NewKeySet flattens a settings map into the dotted keys it defines.
func NewKeySet(settings map[string]any) KeySet {
	keys := KeySet{}

	keys.walk("", settings)

	return keys
}

// Has reports whether the file wrote this key.
func (k KeySet) Has(key string) bool {
	_, ok := k[strings.ToLower(key)]

	return ok
}

func (k KeySet) walk(prefix string, value any) {
	if prefix != "" {
		k[strings.ToLower(prefix)] = struct{}{}
	}

	switch typed := value.(type) {
	case map[string]any:
		for name, child := range typed {
			k.walk(join(prefix, name), child)
		}
	case []any:
		for index, child := range typed {
			k.walk(join(prefix, fmt.Sprint(index)), child)
		}
	}
}

func join(prefix, name string) string {
	if prefix == "" {
		return name
	}

	return prefix + "." + name
}

func providerKey(index int, field string) string {
	return fmt.Sprintf("providers.%d.%s", index, field)
}
