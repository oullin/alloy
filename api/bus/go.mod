module github.com/oullin/alloy/bus

go 1.26.4

require github.com/oullin/alloy/queue v0.0.0

require github.com/oullin/alloy/api/contracts v0.0.0 // indirect

replace github.com/oullin/alloy/queue => ../queue

replace github.com/oullin/alloy/api/contracts => ../contracts
