package cache

import (
	"errors"

	ccache "github.com/oullin/alloy/api/contracts/cache"
)

// Store is a TTL-aware cache contract.
type Store = ccache.Store

var ErrNotFound = errors.New("cache: key not found")
