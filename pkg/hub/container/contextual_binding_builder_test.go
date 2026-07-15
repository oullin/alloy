package container_test

import (
	"testing"

	"github.com/oullin/alloy/pkg/hub/container"
)

type configStub struct {
	data map[string]any
}

func TestContextualBindingGiveTagged(t *testing.T) {
	t.Parallel()

	c := newContainer()

	c.Bind("report-a", func(_ *container.App) (any, error) {
		return "report-a", nil
	}, false)

	c.Bind("report-b", func(_ *container.App) (any, error) {
		return "report-b", nil
	}, false)

	c.Tag([]string{"report-a", "report-b"}, "reports")

	c.When("aggregator").Needs("reports").GiveTagged("reports")

	c.Bind("aggregator", func(cc *container.App) (any, error) {
		return cc.Make("reports")
	}, false)

	v, err := c.Make("aggregator")

	if err != nil {
		t.Fatal(err)
	}

	reports, ok := v.([]any)

	if !ok {
		t.Fatalf("expected []any, got %T", v)
	}

	if len(reports) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(reports))
	}

	if reports[0] != "report-a" || reports[1] != "report-b" {
		t.Fatalf("unexpected report values: %v", reports)
	}
}

func TestContextualBindingGiveConfig(t *testing.T) {
	t.Parallel()

	c := newContainer()

	type mockConfig struct{}

	c.Instance("config", &mockConfig{})

	c.When("service").Needs("$timeout").GiveConfig("app.timeout", 30)

	c.Bind("service", func(cc *container.App) (any, error) {
		return cc.Make("$timeout")
	}, false)

	v, _ := c.Make("service")

	if v != 30 {
		t.Fatalf("expected 30, got %v", v)
	}
}

func TestContextualBindingGiveConfigWithGetter(t *testing.T) {
	t.Parallel()

	c := newContainer()

	type mockConfig struct{}

	mc := &struct {
		mockConfig
	}{}

	cfg := &configStub{data: map[string]any{
		"app.timeout": 60,
	}}
	c.Instance("config", cfg)

	_ = mc

	c.When("service").Needs("$timeout").GiveConfig("app.timeout", 30)

	c.Bind("service", func(cc *container.App) (any, error) {
		return cc.Make("$timeout")
	}, false)

	v, _ := c.Make("service")

	if v != 60 {
		t.Fatalf("expected 60, got %v", v)
	}
}

func (cs *configStub) Get(key string, fallback ...any) any {
	if v, ok := cs.data[key]; ok {
		return v
	}

	if len(fallback) > 0 {
		return fallback[0]
	}

	return nil
}
