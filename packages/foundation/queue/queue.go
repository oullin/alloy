package queue

import cqueue "github.com/oullin/alloy/packages/foundation/contracts/queue"

// Backend defines the interface for a queue backend.
type Backend = cqueue.Backend

// Connector creates a Backend from a configuration map.
type Connector = cqueue.Connector
