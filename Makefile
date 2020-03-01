.PHONY: all
all: check

.PHONY: check
check:
	go vet
	golangci-lint run
	go test -v ./...
