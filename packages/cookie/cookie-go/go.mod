module github.com/oullin/alloy/cookie

go 1.26.0

require (
	github.com/oullin/alloy/container v0.0.0
	github.com/oullin/alloy/contracts v0.0.0
)

replace (
	github.com/oullin/alloy/container => ../../container/container-go
	github.com/oullin/alloy/contracts => ../../contracts/contracts-go
)
