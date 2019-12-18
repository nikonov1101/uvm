build: build/uasm build/uvm

build/uasm:
	go build -o uasm ./cmd/asm/

build/uvm:
	go build -o uvm ./cmd/uvm

vet:
	@echo "+ $@"
	@go vet $(shell go list ./...)

fmt:
	@echo "+ $@"
	@test -z "$$(gofmt -s -l . 2>&1 | grep -v ^vendor/ | tee /dev/stderr)" || \
		(echo >&2 "+ please format Go code with 'gofmt -s'" && false)

test:
	@echo "+ $@"
	go test -count=1 -tags nocgo $(shell go list ./...)
