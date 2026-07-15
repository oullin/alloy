package alias_test

import (
	"testing"

	"github.com/oullin/alloy/pkg/hub/container/internal/alias"
)

func TestResolveReturnsUnknownNameUnchanged(t *testing.T) {
	t.Parallel()

	tbl := alias.NewTable()

	if got := tbl.Resolve("nope"); got != "nope" {
		t.Fatalf("got %q, want %q", got, "nope")
	}
}

func TestResolveSingleHop(t *testing.T) {
	t.Parallel()

	tbl := alias.NewTable()
	tbl.Add("db", "database")

	if got := tbl.Resolve("database"); got != "db" {
		t.Fatalf("got %q, want %q", got, "db")
	}
}

func TestResolveWalksChain(t *testing.T) {
	t.Parallel()

	tbl := alias.NewTable()
	tbl.Add("db", "database")
	tbl.Add("database", "conn")

	if got := tbl.Resolve("conn"); got != "db" {
		t.Fatalf("got %q, want %q", got, "db")
	}
}

func TestHas(t *testing.T) {
	t.Parallel()

	tbl := alias.NewTable()
	tbl.Add("db", "database")

	if !tbl.Has("database") {
		t.Fatal("expected the alias to be registered")
	}

	// The abstract itself is not an alias.
	if tbl.Has("db") {
		t.Fatal("the abstract must not report as an alias")
	}
}

// Drop and Remove differ only in whether the reverse index is purged. The
// container relies on that difference (dropStale vs Instance), so pin it.

func TestDropRemovesForwardEntryOnly(t *testing.T) {
	t.Parallel()

	tbl := alias.NewTable()
	tbl.Add("db", "database")

	tbl.Drop("database")

	if tbl.Has("database") {
		t.Fatal("Drop must remove the forward entry")
	}

	if got := tbl.Resolve("database"); got != "database" {
		t.Fatalf("got %q, want the name returned unchanged", got)
	}

	// The reverse index is deliberately left intact: re-adding the abstract
	// and removing it must still find the stale reverse entry to purge.
	tbl.Remove("database")

	if tbl.Has("database") {
		t.Fatal("Remove after Drop must stay clean")
	}
}

func TestRemovePurgesReverseIndex(t *testing.T) {
	t.Parallel()

	tbl := alias.NewTable()
	tbl.Add("db", "database")
	tbl.Add("db", "conn")

	tbl.Remove("database")

	if tbl.Has("database") {
		t.Fatal("Remove must remove the forward entry")
	}

	// The sibling alias against the same abstract must survive.
	if !tbl.Has("conn") {
		t.Fatal("Remove must not disturb other aliases of the same abstract")
	}

	if got := tbl.Resolve("conn"); got != "db" {
		t.Fatalf("got %q, want %q", got, "db")
	}
}

func TestRemoveIsSafeForUnknownName(t *testing.T) {
	t.Parallel()

	tbl := alias.NewTable()
	tbl.Add("db", "database")

	tbl.Remove("never-registered")

	if !tbl.Has("database") {
		t.Fatal("removing an unknown name must not disturb the table")
	}
}

func TestMultipleAliasesForOneAbstract(t *testing.T) {
	t.Parallel()

	tbl := alias.NewTable()
	tbl.Add("db", "database")
	tbl.Add("db", "conn")

	for _, name := range []string{"database", "conn"} {
		if got := tbl.Resolve(name); got != "db" {
			t.Fatalf("Resolve(%q) = %q, want %q", name, got, "db")
		}
	}
}

func TestReset(t *testing.T) {
	t.Parallel()

	tbl := alias.NewTable()
	tbl.Add("db", "database")

	tbl.Reset()

	if tbl.Has("database") {
		t.Fatal("Reset must clear the forward map")
	}

	// Reset must leave the table usable, not just empty.
	tbl.Add("cache", "redis")

	if got := tbl.Resolve("redis"); got != "cache" {
		t.Fatalf("got %q, want %q", got, "cache")
	}
}

// Resolve and Has run under a read lock in App, so they must not mutate the
// table. A hidden write would turn an RLock reader into a writer and race.
func TestReadsDoNotMutate(t *testing.T) {
	t.Parallel()

	tbl := alias.NewTable()
	tbl.Add("db", "database")
	tbl.Add("database", "conn")

	// A deep chain walk is where path-compression would be tempting.
	tbl.Resolve("conn")
	tbl.Resolve("missing")
	tbl.Has("missing")

	// If Resolve had compressed the chain, "conn" would now point straight at
	// "db" and dropping the middle hop would stop mattering.
	tbl.Drop("database")

	if got := tbl.Resolve("conn"); got != "database" {
		t.Fatalf("got %q, want %q — Resolve mutated the table", got, "database")
	}
}
