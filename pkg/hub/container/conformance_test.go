package container_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"hara.sh/alloy/container"
	"hara.sh/alloy/container/contracts/provider"
)

type containerFixture struct {
	SchemaVersion int                    `json:"schemaVersion"`
	Cases         []containerFixtureCase `json:"cases"`
}

type containerFixtureCase struct {
	ID            string               `json:"id"`
	Note          string               `json:"note"`
	TokensRaw     json.RawMessage      `json:"tokens"`
	ProvidersRaw  json.RawMessage      `json:"providers"`
	OperationsRaw json.RawMessage      `json:"operations"`
	Tokens        []containerToken     `json:"-"`
	Providers     []containerProvider  `json:"-"`
	Operations    []containerOperation `json:"-"`
	Expected      []string             `json:"expected"`
	Error         string               `json:"error"`
}

type containerToken struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

type fixtureScalar struct {
	set bool
	val any
}

type containerProviderValue struct {
	Token string        `json:"token"`
	Value fixtureScalar `json:"value"`
}

type containerProvider struct {
	ID              string                  `json:"id"`
	Provides        []string                `json:"provides"`
	DependsOn       []string                `json:"dependsOn"`
	Deferred        bool                    `json:"deferred"`
	RegisterEvent   string                  `json:"registerEvent"`
	BootEvent       string                  `json:"bootEvent"`
	RegisterValue   *containerProviderValue `json:"registerValue"`
	RegisterResolve string                  `json:"registerResolve"`
}

type containerOperation struct {
	Kind       string                   `json:"kind"`
	Token      string                   `json:"token"`
	Target     string                   `json:"target"`
	Alias      string                   `json:"alias"`
	Primitive  string                   `json:"primitive"`
	Lifetime   string                   `json:"lifetime"`
	Counter    string                   `json:"counter"`
	Parameter  string                   `json:"parameter"`
	Value      fixtureScalar            `json:"value"`
	Suffix     string                   `json:"suffix"`
	Parameters map[string]fixtureScalar `json:"parameters"`
	Observe    string                   `json:"observe"`
	Consumer   string                   `json:"consumer"`
	Consumers  []string                 `json:"consumers"`
	Needs      string                   `json:"needs"`
	Tokens     []string                 `json:"tokens"`
	Tag        string                   `json:"tag"`
	Phase      string                   `json:"phase"`
	Event      string                   `json:"event"`
	Method     string                   `json:"method"`
	Instance   fixtureScalar            `json:"instance"`
	Provider   string                   `json:"provider"`
	Providers  []string                 `json:"providers"`
}

type fixtureProvider struct {
	spec   containerProvider
	app    *container.Application
	events *[]string
}

func (s *fixtureScalar) UnmarshalJSON(data []byte) error {
	s.set = true

	if string(data) == "null" {
		s.val = nil

		return nil
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	var value any

	if err := dec.Decode(&value); err != nil {
		return err
	}

	switch value.(type) {
	case string, bool, json.Number:
		s.val = value

		return nil
	default:
		return fmt.Errorf("must be a scalar")
	}
}

func (s fixtureScalar) value() any {
	if !s.set {
		return nil
	}

	return s.val
}

var fixtureKinds = map[string]bool{
	"bind": true, "bind-if": true, "singleton-if": true, "scoped-if": true,
	"instance": true, "resolve": true, "resolve-with-parameters": true, "get": true,
	"forget-scoped": true, "forget-instance": true, "flush": true, "alias": true,
	"contextual-value": true, "contextual-factory": true, "contextual-tagged": true,
	"tag": true, "tagged": true, "extend": true, "callback": true, "rebinding": true,
	"method-bind": true, "method-call": true, "call": true, "wrap": true, "factory-func": true,
	"provider-register": true, "provider-register-many": true, "provider-boot": true,
	"observe-counter": true, "observe-events": true, "observe-bound": true, "observe-has": true,
	"observe-resolved": true, "observe-is-shared": true, "observe-bindings": true,
	"observe-providers": true, "observe-has-provider": true, "observe-provider-for": true,
	"observe-booted": true, "observe-has-method": true,
}

var fixturePrimitives = map[string]bool{
	"constant": true, "increment-counter": true, "resolve-token": true,
	"read-parameter": true, "append-suffix": true, "return-instance": true,
}

var fixtureErrors = map[string]bool{
	"ALIAS_CYCLE": true, "MISSING_BINDING": true, "CIRCULAR_RESOLUTION": true,
	"SELF_ALIAS": true, "MISSING_METHOD_BINDING": true, "PROVIDER_CYCLE": true,
}

var fixtureLifetimes = map[string]bool{"transient": true, "singleton": true, "scoped": true}

// TestContainerConformance runs the same ordered DSL operations and observations as sdk/container.
func TestContainerConformance(t *testing.T) {
	t.Parallel()
	fixture := loadContainerConformance(t)

	for _, tc := range fixture.Cases {
		t.Run(tc.ID, func(t *testing.T) {
			got, err := runContainerCase(tc)

			if tc.Error != "" {
				if !containerConformanceError(tc.Error, err) {
					t.Fatalf("error = %v, want %s (%s)", err, tc.Error, tc.Note)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v (%s)", err, tc.Note)
			}

			if !slices.Equal(got, tc.Expected) {
				t.Fatalf("observations = %q, want %q (%s)", got, tc.Expected, tc.Note)
			}
		})
	}
}

func TestContainerConformanceFixtureValidation(t *testing.T) {
	valid := `{"schemaVersion":1,"cases":[{"id":"valid","note":"valid fixture","tokens":[{"id":"token"}],"operations":[{"kind":"resolve","token":"token"}],"error":"MISSING_BINDING"}]}`

	for _, tc := range []struct{ name, fixture string }{
		{"unknown operation", strings.Replace(valid, `"resolve"`, `"unknown"`, 1)},
		{"unknown error identity", strings.Replace(valid, `"MISSING_BINDING"`, `"UNKNOWN"`, 1)},
		{"both expected and error", strings.Replace(valid, `"error":"MISSING_BINDING"`, `"expected":[],"error":"MISSING_BINDING"`, 1)},
		{"neither expected nor error", strings.Replace(valid, `,"error":"MISSING_BINDING"`, "", 1)},
		{"duplicate case id", `{"schemaVersion":1,"cases":[{"id":"same","note":"a","tokens":[],"operations":[],"error":"MISSING_BINDING"},{"id":"same","note":"b","tokens":[],"operations":[],"error":"MISSING_BINDING"}]}`},
		{"missing note", strings.Replace(valid, `"note":"valid fixture",`, "", 1)},
		{"invalid token reference", strings.Replace(valid, `"token":"token"`, `"token":"missing"`, 1)},
		{"missing tokens array", `{"schemaVersion":1,"cases":[{"id":"valid","note":"valid fixture","operations":[],"error":"MISSING_BINDING"}]}`},
		{"null tokens array", `{"schemaVersion":1,"cases":[{"id":"valid","note":"valid fixture","tokens":null,"operations":[],"error":"MISSING_BINDING"}]}`},
		{"missing operations array", `{"schemaVersion":1,"cases":[{"id":"valid","note":"valid fixture","tokens":[],"error":"MISSING_BINDING"}]}`},
		{"null operations array", `{"schemaVersion":1,"cases":[{"id":"valid","note":"valid fixture","tokens":[],"operations":null,"error":"MISSING_BINDING"}]}`},
		{"invalid lifetime", strings.Replace(valid, `"operations":[{"kind":"resolve","token":"token"}]`, `"operations":[{"kind":"bind","token":"token","lifetime":"request","primitive":"constant","value":"v"}]`, 1)},
		{"rebinding without event", strings.Replace(valid, `"operations":[{"kind":"resolve","token":"token"}]`, `"operations":[{"kind":"rebinding","token":"token"}]`, 1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := parseContainerFixture([]byte(tc.fixture)); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func containerConformanceError(identity string, err error) bool {
	switch identity {
	case "ALIAS_CYCLE":
		return errors.Is(err, container.ErrAliasCycle)
	case "MISSING_BINDING":
		return errors.Is(err, container.ErrNotBound)
	case "CIRCULAR_RESOLUTION":
		return errors.Is(err, container.ErrCircularDependency)
	case "SELF_ALIAS":
		return errors.Is(err, container.ErrSelfAlias)
	case "MISSING_METHOD_BINDING":
		return errors.Is(err, container.ErrMethodNotBound)
	case "PROVIDER_CYCLE":
		return errors.Is(err, container.ErrProviderCycle)
	}

	return false
}

func (p *fixtureProvider) Register() {
	if p.spec.RegisterEvent != "" {
		*p.events = append(*p.events, p.spec.RegisterEvent)
	}

	if p.spec.RegisterResolve != "" {
		_, _ = p.app.Make(p.spec.RegisterResolve)
	}

	if p.spec.RegisterValue != nil {
		p.app.Instance(p.spec.RegisterValue.Token, p.spec.RegisterValue.Value.value())
	}
}

func (p *fixtureProvider) Boot() {
	if p.spec.BootEvent != "" {
		*p.events = append(*p.events, p.spec.BootEvent)
	}
}

func (p *fixtureProvider) Provides() []string { return p.spec.Provides }

func (p *fixtureProvider) DependsOn() []string { return p.spec.DependsOn }

func (p *fixtureProvider) Deferred() bool { return p.spec.Deferred }

func runContainerCase(tc containerFixtureCase) (observations []string, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if recoveredErr, ok := recovered.(error); ok && errors.Is(recoveredErr, container.ErrProviderCycle) {
				err = recoveredErr

				return
			}

			panic(recovered)
		}
	}()

	app := container.NewApplication()
	counters := map[string]int{}
	events := []string{}
	providers := map[string]*fixtureProvider{}
	providerIDs := map[provider.ServiceProvider]string{}

	for _, spec := range tc.Providers {
		fp := &fixtureProvider{spec: spec, app: app, events: &events}
		providers[spec.ID] = fp
		providerIDs[fp] = spec.ID
	}

	runPrimitive := func(op containerOperation, c *container.App, params map[string]any) (any, error) {
		switch op.Primitive {
		case "constant":
			return op.Value.value(), nil
		case "increment-counter":
			counters[op.Counter]++

			return counters[op.Counter], nil
		case "resolve-token":
			return c.Make(op.Target)
		case "read-parameter":
			if params == nil {
				return "<nil>", nil
			}

			value, ok := params[op.Parameter]

			if !ok {
				return "<nil>", nil
			}

			return value, nil
		case "return-instance":
			if params == nil {
				return nil, nil
			}

			return params["_instance"], nil
		default:
			return nil, fmt.Errorf("unknown factory primitive %q", op.Primitive)
		}
	}

	factory := func(op containerOperation) container.Factory {
		return func(c *container.App) (any, error) {
			return runPrimitive(op, c, c.Parameters())
		}
	}

	methodCallable := func(op containerOperation) container.MethodCallable {
		return func(c *container.App, params map[string]any) (any, error) {
			return runPrimitive(op, c, params)
		}
	}

	extender := func(op containerOperation) container.ExtenderFunc {
		return func(value any, _ *container.App) (any, error) {
			if op.Primitive != "append-suffix" {
				return value, nil
			}

			return renderFixtureValue(value) + op.Suffix, nil
		}
	}

	scalarParams := func(values map[string]fixtureScalar) map[string]any {
		parameters := make(map[string]any, len(values))

		for key, value := range values {
			parameters[key] = value.value()
		}

		return parameters
	}

	observe := func(op containerOperation, value any) {
		if op.Observe != "" {
			observations = append(observations, op.Observe+"="+renderFixtureValue(value))
		}
	}

	bindWithLifetime := func(token string, op containerOperation, conditional bool) {
		switch op.Lifetime {
		case "singleton":
			if conditional {
				app.SingletonIf(token, factory(op))
			} else {
				app.Singleton(token, factory(op))
			}
		case "scoped":
			if conditional {
				app.ScopedIf(token, factory(op))
			} else {
				app.Scoped(token, factory(op))
			}
		default:
			if conditional {
				app.BindIf(token, factory(op), false)
			} else {
				app.Bind(token, factory(op), false)
			}
		}
	}

	contextualConsumers := func(op containerOperation) []string {
		if len(op.Consumers) > 0 {
			return op.Consumers
		}

		return []string{op.Consumer}
	}

	for _, op := range tc.Operations {
		switch op.Kind {
		case "bind":
			bindWithLifetime(op.Token, op, false)
		case "bind-if":
			bindWithLifetime(op.Token, op, true)
		case "singleton-if":
			app.SingletonIf(op.Token, factory(op))
		case "scoped-if":
			app.ScopedIf(op.Token, factory(op))
		case "instance":
			app.Instance(op.Token, op.Value.value())
		case "resolve":
			value, resolveErr := app.Make(op.Token)

			if resolveErr != nil {
				return observations, resolveErr
			}

			observe(op, value)
		case "resolve-with-parameters":
			value, resolveErr := app.MakeWith(op.Token, scalarParams(op.Parameters))

			if resolveErr != nil {
				return observations, resolveErr
			}

			observe(op, value)
		case "get":
			value, getErr := app.Get(op.Token)

			if getErr != nil {
				return observations, getErr
			}

			observe(op, value)
		case "forget-scoped":
			app.ForgetScopedInstances()
		case "forget-instance":
			app.ForgetInstance(op.Token)
		case "flush":
			app.Flush()
		case "alias":
			if aliasErr := app.Alias(op.Target, op.Alias); aliasErr != nil {
				return observations, aliasErr
			}
		case "contextual-value":
			app.When(contextualConsumers(op)...).Needs(op.Needs).Give(op.Value.value())
		case "contextual-factory":
			app.When(contextualConsumers(op)...).Needs(op.Needs).Give(factory(op))
		case "contextual-tagged":
			app.When(contextualConsumers(op)...).Needs(op.Needs).GiveTagged(op.Tag)
		case "tag":
			app.Tag(op.Tokens, op.Tag)
		case "tagged":
			observe(op, app.Tagged(op.Tag))
		case "extend":
			app.Extend(op.Token, extender(op))
		case "callback":
			switch op.Phase {
			case "before":
				if op.Token == "" {
					app.BeforeResolvingAny(func(string, map[string]any, *container.App) { events = append(events, op.Event) })
				} else {
					app.BeforeResolving(op.Token, func(string, map[string]any, *container.App) { events = append(events, op.Event) })
				}
			case "resolving":
				if op.Token == "" {
					app.ResolvingAny(func(any, *container.App) { events = append(events, op.Event) })
				} else {
					app.Resolving(op.Token, func(any, *container.App) { events = append(events, op.Event) })
				}
			case "after":
				if op.Token == "" {
					app.AfterResolvingAny(func(any, *container.App) { events = append(events, op.Event) })
				} else {
					app.AfterResolving(op.Token, func(any, *container.App) { events = append(events, op.Event) })
				}
			}
		case "rebinding":
			_, rebindingErr := app.Rebinding(op.Token, func(value any, _ *container.App) {
				events = append(events, op.Event+":"+renderFixtureValue(value))
			})

			if rebindingErr != nil {
				return observations, rebindingErr
			}
		case "method-bind":
			app.BindMethod(op.Method, methodCallable(op))
		case "method-call":
			value, callErr := app.CallMethodBinding(op.Method, op.Instance.value())

			if callErr != nil {
				return observations, callErr
			}

			observe(op, value)
		case "call":
			value, callErr := app.Call(methodCallable(op), scalarParams(op.Parameters))

			if callErr != nil {
				return observations, callErr
			}

			observe(op, value)
		case "wrap":
			value, wrapErr := app.Wrap(methodCallable(op), scalarParams(op.Parameters))()

			if wrapErr != nil {
				return observations, wrapErr
			}

			observe(op, value)
		case "factory-func":
			value, factoryErr := app.FactoryFunc(op.Token)()

			if factoryErr != nil {
				return observations, factoryErr
			}

			observe(op, value)
		case "provider-register":
			app.Register(providers[op.Provider])
		case "provider-register-many":
			values := make([]provider.ServiceProvider, 0, len(op.Providers))

			for _, id := range op.Providers {
				values = append(values, providers[id])
			}

			app.RegisterMany(values)
		case "provider-boot":
			app.Boot()
		case "observe-counter":
			observations = append(observations, op.Counter+"="+fmt.Sprint(counters[op.Counter]))
		case "observe-events":
			name := op.Observe

			if name == "" {
				name = "events"
			}

			observations = append(observations, name+"="+strings.Join(events, ","))
		case "observe-bound":
			observe(op, app.Bound(op.Token))
		case "observe-has":
			observe(op, app.Has(op.Token))
		case "observe-resolved":
			observe(op, app.Resolved(op.Token))
		case "observe-is-shared":
			observe(op, app.IsShared(op.Token))
		case "observe-bindings":
			bindings := app.GetBindings()
			keys := make([]string, 0, len(bindings))

			for key := range bindings {
				keys = append(keys, key)
			}

			sort.Strings(keys)

			parts := make([]string, 0, len(keys))

			for _, key := range keys {
				binding := bindings[key]
				lifetime := "transient"

				if binding.Scoped() {
					lifetime = "scoped"
				} else if binding.Shared() {
					lifetime = "singleton"
				}

				parts = append(parts, key+":"+lifetime)
			}

			observe(op, strings.Join(parts, ","))
		case "observe-providers":
			ids := make([]string, 0, len(app.Providers()))

			for _, registered := range app.Providers() {
				ids = append(ids, providerIDs[registered])
			}

			observe(op, strings.Join(ids, ","))
		case "observe-has-provider":
			observe(op, app.HasProvider(op.Token))
		case "observe-provider-for":
			found := app.ProviderFor(op.Token)

			if found == nil {
				observe(op, nil)
			} else {
				observe(op, providerIDs[found])
			}
		case "observe-booted":
			observe(op, app.Booted())
		case "observe-has-method":
			observe(op, app.HasMethodBinding(op.Method))
		}
	}

	return observations, nil
}

func renderFixtureValue(value any) string {
	if value == nil {
		return "<nil>"
	}

	switch typed := value.(type) {
	case []any:
		rendered := make([]string, 0, len(typed))

		for _, item := range typed {
			rendered = append(rendered, renderFixtureValue(item))
		}

		return strings.Join(rendered, ",")
	case json.Number:
		return typed.String()
	case bool:
		return strconv.FormatBool(typed)
	case string:
		return typed
	default:
		return fmt.Sprint(value)
	}
}

func loadContainerConformance(t *testing.T) containerFixture {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)

	if !ok {
		t.Fatal("cannot resolve conformance test path")
	}

	data, err := os.ReadFile(filepath.Join(filepath.Dir(file), "..", "..", "..", "conformance", "container.json"))

	if err != nil {
		t.Fatal(err)
	}

	fixture, err := parseContainerFixture(data)

	if err != nil {
		t.Fatal(err)
	}

	return fixture
}

func parseContainerFixture(data []byte) (containerFixture, error) {
	var fixture containerFixture

	if err := json.Unmarshal(data, &fixture); err != nil {
		return containerFixture{}, err
	}

	if fixture.SchemaVersion != 1 {
		return containerFixture{}, fmt.Errorf("fixture.schemaVersion must be 1")
	}

	if len(fixture.Cases) == 0 {
		return containerFixture{}, fmt.Errorf("fixture.cases must be non-empty")
	}

	ids := map[string]bool{}

	for index := range fixture.Cases {
		tc := &fixture.Cases[index]
		where := fmt.Sprintf("fixture.cases[%d]", index)

		if tc.ID == "" || ids[tc.ID] {
			return containerFixture{}, fmt.Errorf("%s has empty or duplicate id", where)
		}

		ids[tc.ID] = true

		if tc.Note == "" {
			return containerFixture{}, fmt.Errorf("%s.note must be non-empty", where)
		}

		if (tc.Expected == nil) == (tc.Error == "") {
			return containerFixture{}, fmt.Errorf("%s must declare exactly one of expected or error", where)
		}

		if tc.Error != "" && !fixtureErrors[tc.Error] {
			return containerFixture{}, fmt.Errorf("%s.error is unknown", where)
		}

		if err := decodeRequiredArray(tc.TokensRaw, &tc.Tokens, where+".tokens"); err != nil {
			return containerFixture{}, err
		}

		if tc.ProvidersRaw == nil {
			tc.Providers = nil
		} else if string(tc.ProvidersRaw) == "null" {
			return containerFixture{}, fmt.Errorf("%s.providers must be an array", where)
		} else if err := json.Unmarshal(tc.ProvidersRaw, &tc.Providers); err != nil {
			return containerFixture{}, fmt.Errorf("%s.providers: %w", where, err)
		}

		if err := decodeRequiredArray(tc.OperationsRaw, &tc.Operations, where+".operations"); err != nil {
			return containerFixture{}, err
		}

		if err := validateContainerCase(*tc, where); err != nil {
			return containerFixture{}, err
		}
	}

	return fixture, nil
}

func decodeRequiredArray[T any](raw json.RawMessage, dest *[]T, where string) error {
	if len(raw) == 0 || string(raw) == "null" {
		return fmt.Errorf("%s must be an array", where)
	}

	return json.Unmarshal(raw, dest)
}

func validateContainerCase(tc containerFixtureCase, where string) error {
	tokens := map[string]bool{}

	for _, token := range tc.Tokens {
		if token.ID == "" || tokens[token.ID] {
			return fmt.Errorf("%s has empty or duplicate token", where)
		}

		if token.Kind != "" && token.Kind != "string" {
			return fmt.Errorf("%s token %q has unknown kind", where, token.ID)
		}

		tokens[token.ID] = true
	}

	providers := map[string]bool{}

	for _, spec := range tc.Providers {
		if spec.ID == "" || providers[spec.ID] {
			return fmt.Errorf("%s has empty or duplicate provider", where)
		}

		providers[spec.ID] = true

		for _, key := range append(slices.Clone(spec.Provides), spec.DependsOn...) {
			if !tokens[key] {
				return fmt.Errorf("%s provider %q references unknown token %q", where, spec.ID, key)
			}
		}

		if spec.RegisterValue != nil {
			if !tokens[spec.RegisterValue.Token] {
				return fmt.Errorf("%s provider %q references unknown token", where, spec.ID)
			}

			if !spec.RegisterValue.Value.set {
				return fmt.Errorf("%s provider %q registerValue.value is required", where, spec.ID)
			}
		}

		if spec.RegisterResolve != "" && !tokens[spec.RegisterResolve] {
			return fmt.Errorf("%s provider %q registerResolve references unknown token", where, spec.ID)
		}
	}

	for index, op := range tc.Operations {
		if !fixtureKinds[op.Kind] {
			return fmt.Errorf("%s.operations[%d].kind is unknown", where, index)
		}

		if op.Primitive != "" && !fixturePrimitives[op.Primitive] {
			return fmt.Errorf("%s.operations[%d].primitive is unknown", where, index)
		}

		if op.Lifetime != "" && !fixtureLifetimes[op.Lifetime] {
			return fmt.Errorf("%s.operations[%d].lifetime is unknown", where, index)
		}

		if err := validateOperation(op, tokens, providers); err != nil {
			return fmt.Errorf("%s.operations[%d]: %w", where, index, err)
		}
	}

	return nil
}

func validateOperation(op containerOperation, tokens, providers map[string]bool) error {
	token := func(value string) error {
		if value == "" || !tokens[value] {
			return fmt.Errorf("unknown token %q", value)
		}

		return nil
	}

	requirePrimitive := func() error {
		if op.Primitive == "" {
			return errors.New("primitive is required")
		}

		return nil
	}

	require := func(value, field string) error {
		if value == "" {
			return fmt.Errorf("%s is required", field)
		}

		return nil
	}

	requireScalar := func(value fixtureScalar, field string) error {
		if !value.set {
			return fmt.Errorf("%s is required", field)
		}

		return nil
	}

	requireConsumers := func() error {
		if op.Consumer != "" && len(op.Consumers) > 0 {
			return errors.New("ambiguous consumer-versus-consumers representation")
		}

		if op.Consumer == "" && len(op.Consumers) == 0 {
			return errors.New("consumer is required")
		}

		return nil
	}

	switch op.Kind {
	case "bind", "bind-if", "singleton-if", "scoped-if", "extend", "method-bind", "contextual-factory", "call", "wrap":
		if err := requirePrimitive(); err != nil {
			return err
		}
	}

	switch op.Kind {
	case "bind", "bind-if", "singleton-if", "scoped-if", "resolve", "get", "forget-instance", "extend", "factory-func",
		"observe-bound", "observe-has", "observe-resolved", "observe-is-shared", "observe-has-provider", "observe-provider-for":
		if err := require(op.Token, "token"); err != nil {
			return err
		}
	case "resolve-with-parameters":
		if err := require(op.Token, "token"); err != nil {
			return err
		}

		if op.Parameters == nil {
			return errors.New("parameters is required")
		}
	case "instance":
		if err := require(op.Token, "token"); err != nil {
			return err
		}

		if err := requireScalar(op.Value, "value"); err != nil {
			return err
		}
	case "contextual-value":
		if err := requireConsumers(); err != nil {
			return err
		}

		if err := require(op.Needs, "needs"); err != nil {
			return err
		}

		if err := requireScalar(op.Value, "value"); err != nil {
			return err
		}
	case "alias":
		if err := require(op.Target, "target"); err != nil {
			return err
		}

		if err := require(op.Alias, "alias"); err != nil {
			return err
		}
	case "contextual-factory", "contextual-tagged":
		if err := requireConsumers(); err != nil {
			return err
		}

		if err := require(op.Needs, "needs"); err != nil {
			return err
		}

		if op.Kind == "contextual-tagged" {
			if err := require(op.Tag, "tag"); err != nil {
				return err
			}
		}
	case "tag":
		if err := require(op.Tag, "tag"); err != nil {
			return err
		}

		if op.Tokens == nil {
			return errors.New("tokens is required")
		}
	case "tagged":
		if err := require(op.Tag, "tag"); err != nil {
			return err
		}
	case "callback":
		if err := require(op.Phase, "phase"); err != nil {
			return err
		}

		if err := require(op.Event, "event"); err != nil {
			return err
		}
	case "rebinding":
		if err := require(op.Token, "token"); err != nil {
			return err
		}

		if err := require(op.Event, "event"); err != nil {
			return err
		}
	case "method-bind", "method-call", "observe-has-method":
		if err := require(op.Method, "method"); err != nil {
			return err
		}

		if op.Kind == "method-call" {
			if err := requireScalar(op.Instance, "instance"); err != nil {
				return err
			}
		}
	case "provider-register":
		if err := require(op.Provider, "provider"); err != nil {
			return err
		}
	case "provider-register-many":
		if op.Providers == nil {
			return errors.New("providers is required")
		}
	case "observe-counter":
		if err := require(op.Counter, "counter"); err != nil {
			return err
		}
	}

	if op.Primitive == "constant" && !op.Value.set {
		return errors.New("constant primitive requires value")
	}

	if op.Primitive == "increment-counter" && op.Counter == "" {
		return errors.New("increment-counter primitive requires counter")
	}

	if op.Primitive == "resolve-token" && op.Target == "" {
		return errors.New("resolve-token primitive requires target")
	}

	if op.Primitive == "read-parameter" && op.Parameter == "" {
		return errors.New("read-parameter primitive requires parameter")
	}

	if op.Primitive == "append-suffix" && op.Suffix == "" {
		return errors.New("append-suffix primitive requires suffix")
	}

	if op.Kind == "contextual-value" && op.Primitive != "" {
		return errors.New("ambiguous value-versus-factory representation")
	}

	switch op.Kind {
	case "bind", "bind-if", "singleton-if", "scoped-if", "instance", "resolve", "resolve-with-parameters", "get",
		"forget-instance", "extend", "rebinding", "factory-func", "observe-bound", "observe-has",
		"observe-resolved", "observe-is-shared", "observe-has-provider", "observe-provider-for":
		if err := token(op.Token); err != nil {
			return err
		}
	case "alias":
		if err := token(op.Target); err != nil {
			return err
		}

		if err := token(op.Alias); err != nil {
			return err
		}
	case "contextual-value", "contextual-factory", "contextual-tagged":
		for _, consumer := range contextualConsumerList(op) {
			if err := token(consumer); err != nil {
				return err
			}
		}

		if err := token(op.Needs); err != nil {
			return err
		}
	case "tag":
		for _, value := range op.Tokens {
			if err := token(value); err != nil {
				return err
			}
		}
	case "callback":
		if op.Token != "" {
			if err := token(op.Token); err != nil {
				return err
			}
		}
	case "provider-register":
		if !providers[op.Provider] {
			return fmt.Errorf("unknown provider %q", op.Provider)
		}
	case "provider-register-many":
		for _, value := range op.Providers {
			if !providers[value] {
				return fmt.Errorf("unknown provider %q", value)
			}
		}
	}

	if op.Primitive == "resolve-token" {
		if err := token(op.Target); err != nil {
			return err
		}
	}

	if op.Kind == "callback" && op.Phase != "before" && op.Phase != "resolving" && op.Phase != "after" {
		return fmt.Errorf("unknown callback phase %q", op.Phase)
	}

	return nil
}

func contextualConsumerList(op containerOperation) []string {
	if len(op.Consumers) > 0 {
		return op.Consumers
	}

	if op.Consumer == "" {
		return nil
	}

	return []string{op.Consumer}
}
