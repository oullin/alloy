module github.com/oullin/alloy/filesystem

go 1.26.0

require (
	github.com/oullin/alloy/container v0.0.0
	github.com/oullin/alloy/contracts v0.0.0
	golang.org/x/sys v0.43.0
)

replace (
	github.com/oullin/alloy/contracts => ../../contracts/contracts-go
	github.com/oullin/alloy/container => ../../container/container-go
)
