.PHONY: all build fmt test dep

build:
	go build

fmt:
	go fmt ./...

test:
	go test -cover github.com/itkq/kinesis-agent-go/...

dep:
	dep ensure
