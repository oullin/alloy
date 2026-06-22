module github.com/oullin/alloy/hashing

go 1.26.0

replace (
	github.com/oullin/alloy/container => ../../container/container-go
	github.com/oullin/alloy/contracts => ../../contracts/contracts-go
)

require (
	github.com/oullin/alloy/container v0.0.0
	github.com/oullin/alloy/contracts v0.0.0
	golang.org/x/crypto v0.50.0
)

require golang.org/x/sys v0.43.0 // indirect
