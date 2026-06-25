module github.com/oullin/alloy/hashing

go 1.26.4

require golang.org/x/crypto v0.52.0

require (
	github.com/oullin/alloy/api/contracts v0.0.0
	golang.org/x/sys v0.45.0 // indirect
)

replace github.com/oullin/alloy/api/contracts => ../contracts
