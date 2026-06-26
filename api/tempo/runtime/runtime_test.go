package runtime_test

import (
	"strings"
	"testing"

	"github.com/oullin/alloy/api/tempo"
)

type mapTranslator map[string]string

func (translator mapTranslator) Message(key string) (any, bool) {
	value, ok := translator[key]

	return value, ok
}

func (translator mapTranslator) Translate(key string, replacements map[string]string) (string, bool) {
	value, ok := translator[key]

	if !ok {
		return "", false
	}

	for name, replacement := range replacements {
		value = strings.ReplaceAll(value, ":"+name, replacement)
	}

	return value, true
}

func assertEqual(t *testing.T, label string, got string, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("%s = %q, want %q", label, got, want)
	}
}

func TestRuntimeScopedTranslator(t *testing.T) {
	firstRuntime := tempo.NewRuntime(
		tempo.RuntimeLocale("en-US"),
		tempo.RuntimeTranslator(mapTranslator{"greeting": "Hello :name"}),
	)
	secondRuntime := tempo.NewRuntime(
		tempo.RuntimeLocale("es-ES"),
		tempo.RuntimeTranslator(mapTranslator{"greeting": "Hola :name"}),
	)

	firstFactory, err := tempo.NewFactory(tempo.WithTimezone("UTC"), tempo.WithRuntime(firstRuntime))

	if err != nil {
		t.Fatalf("new first factory: %v", err)
	}

	secondFactory, err := tempo.NewFactory(tempo.WithTimezone("UTC"), tempo.WithRuntime(secondRuntime))

	if err != nil {
		t.Fatalf("new second factory: %v", err)
	}

	first, err := firstFactory.Parse("2024-05-15")

	if err != nil {
		t.Fatalf("first parse: %v", err)
	}

	second, err := secondFactory.Parse("2024-05-15")

	if err != nil {
		t.Fatalf("second parse: %v", err)
	}

	assertEqual(t, "first Translate()", first.Translate("greeting", map[string]string{"name": "Time"}), "Hello Time")
	assertEqual(t, "second Translate()", second.Translate("greeting", map[string]string{"name": "Time"}), "Hola Time")

	if value, ok := first.TranslationMessage("locale"); !ok || value != "en-US" {
		t.Fatalf("first locale message = %v, %v, want en-US, true", value, ok)
	}

	if value, ok := second.TranslationMessage("locale"); !ok || value != "es-ES" {
		t.Fatalf("second locale message = %v, %v, want es-ES, true", value, ok)
	}

	if !first.HasTranslator() {
		t.Fatalf("HasTranslator() = false, want true")
	}

	assertEqual(t, "Clone().Translate()", first.Clone().Translate("greeting", map[string]string{"name": "Time"}), "Hello Time")
	assertEqual(t, "AddDays().Translate()", first.AddDays(1).Translate("greeting", map[string]string{"name": "Time"}), "Hello Time")
	assertEqual(t, "Mutable AddDays().Translate()", first.Mutable().AddDays(1).Translate("greeting", map[string]string{"name": "Time"}), "Hello Time")

	replaced := first.WithTranslator(mapTranslator{"greeting": "Salut :name"})
	assertEqual(t, "replaced Translate()", replaced.Translate("greeting", map[string]string{"name": "Time"}), "Salut Time")
	assertEqual(t, "original Translate()", first.Translate("greeting", map[string]string{"name": "Time"}), "Hello Time")
}
