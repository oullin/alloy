package filesystem_test

import (
	"reflect"
	"testing"

	contract "hara.sh/alloy/contracts/filesystem"
	"hara.sh/alloy/filesystem"
)

// TestContractCoversLocal catches the drift the compile-time assertion in
// filesystem.go cannot. That assertion proves Local satisfies Filesystem, so it
// fires when the interface gains a method the type lacks — but says nothing
// when Local grows a method the interface never hears about. That is the
// direction this package actually drifted: the contract sat unimported with
// nothing binding it to anything, and its parity with Local survived on luck.
//
// Comparing method sets is cheap and names whatever is missing.
func TestContractCoversLocal(t *testing.T) {
	contractType := reflect.TypeOf((*contract.Filesystem)(nil)).Elem()
	localType := reflect.TypeOf(&filesystem.Local{})

	declared := make(map[string]bool, contractType.NumMethod())

	for i := range contractType.NumMethod() {
		declared[contractType.Method(i).Name] = true
	}

	var missing []string

	for i := range localType.NumMethod() {
		name := localType.Method(i).Name

		if !declared[name] {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		t.Errorf(
			"*Local has %d exported methods the Filesystem contract does not declare: %v\n"+
				"Add them to contracts/filesystem, or drop them from Local.",
			len(missing), missing,
		)
	}
}
