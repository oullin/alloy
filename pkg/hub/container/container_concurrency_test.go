package container_test

import (
	"testing"

	"github.com/oullin/alloy/pkg/hub/container"
)

func TestConcurrentAccess(t *testing.T) {
	t.Parallel()

	c := newContainer()

	done := make(chan bool, 40)

	for i := 0; i < 10; i++ {
		go func() {
			c.Bind("name", func(_ *container.App) (any, error) {
				return "Taylor", nil
			}, false)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			c.Make("name") //nolint:errcheck
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			c.Bound("name")
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			c.Has("name")
			done <- true
		}()
	}

	for i := 0; i < 40; i++ {
		<-done
	}
}
