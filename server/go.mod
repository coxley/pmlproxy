module github.com/coxley/pmlproxy/example

go 1.17

replace github.com/coxley/pmlproxy/pb => /home/coxley/projects/pmlproxy/pb

require (
	github.com/coxley/pmlproxy/pb v0.0.0-00010101000000-000000000000
	github.com/golang/glog v1.0.0
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e
	github.com/prometheus/client_golang v1.12.1
	google.golang.org/grpc v1.45.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5 // indirect
	golang.org/x/sys v0.0.0-20220114195835-da31bd327af9 // indirect
	golang.org/x/text v0.3.6 // indirect
	google.golang.org/genproto v0.0.0-20200825200019-8632dd797987 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)
