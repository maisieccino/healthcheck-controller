.PHONY: run run_debug_remote image okteto
hc-controller:
	go build -o hc-controller cmd/hc-controller/main.go
image:
	docker build \
		-t hc-controller:local .
run:
	go run cmd/hc-controller/main.go
run_debug_remote:
	dlv debug --headless --listen=:2345 --api-version=2 cmd/hc-controller/main.go
