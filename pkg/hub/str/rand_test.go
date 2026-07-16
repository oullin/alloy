package str

import (
	"crypto/rand"
	"errors"
	"strings"
	"testing"
)

// failingRandReader always fails, simulating a CSPRNG failure.
type failingRandReader struct{}

func (failingRandReader) Read([]byte) (int, error) {
	return 0, errors.New("rng failure")
}

// TestGenerateRandomPanicsOnRandFailure asserts that an RNG failure surfaces as
// an explicit CSPRNG panic rather than a nil-pointer dereference from the
// previously-ignored rand.Int error.
func TestGenerateRandomPanicsOnRandFailure(t *testing.T) {
	orig := rand.Reader
	rand.Reader = failingRandReader{}

	defer func() { rand.Reader = orig }()

	defer func() {
		r := recover()

		if r == nil {
			t.Fatal("generateRandom: expected panic on CSPRNG failure, got none")
		}

		err, ok := r.(error)

		if !ok || !strings.Contains(err.Error(), "generate random") {
			t.Fatalf("expected explicit CSPRNG failure, got %v (%T)", r, r)
		}
	}()

	_ = generateRandom(8)
}
