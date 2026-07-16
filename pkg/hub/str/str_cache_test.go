package str

import (
	"sync"
	"testing"
)

// TestRegexCacheConcurrent exercises the pattern-derived regex cache from many
// goroutines to ensure the sync.Map-backed cache is race-safe and yields correct
// results under concurrent load. Run with -race.
func TestRegexCacheConcurrent(t *testing.T) {
	FlushCache()

	patterns := []string{"foo*", "bar?", "a.c", "^[0-9]+$", "*"}
	values := []string{"foobar", "bark", "abc", "12345", "anything"}

	var wg sync.WaitGroup

	for i := 0; i < 64; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for j := 0; j < 500; j++ {
				p := patterns[j%len(patterns)]
				v := values[j%len(values)]

				_ = Is(p, v)
				_ = IsMatch([]string{p}, v)
				_ = Match(p, v)
				_ = MatchAll(p, v)
			}
		}()
	}

	wg.Wait()

	// Correctness must hold after concurrent access.
	if !Is("foo*", "foobar") {
		t.Fatal(`Is("foo*", "foobar") = false, want true`)
	}

	if Is("foo*", "nope") {
		t.Fatal(`Is("foo*", "nope") = true, want false`)
	}
}

func BenchmarkIsUuid(b *testing.B) {
	const id = "550e8400-e29b-41d4-a716-446655440000"

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = IsUuid(id)
	}
}

func BenchmarkSlug(b *testing.B) {
	const title = "Hello World, This Is A Test Title!"

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = Slug(title)
	}
}

func BenchmarkIsGlob(b *testing.B) {
	const pattern = "foo/*/bar?"
	const value = "foo/baz/barX"

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = Is(pattern, value)
	}
}
