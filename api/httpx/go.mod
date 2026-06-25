module github.com/oullin/alloy/httpx

go 1.26.4

require (
	github.com/oullin/alloy/api/contracts v0.0.0
	github.com/oullin/alloy/routing v0.0.0
)

replace github.com/oullin/alloy/routing => ../routing

replace github.com/oullin/alloy/api/contracts => ../contracts
