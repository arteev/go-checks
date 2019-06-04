.PHONY: all

all: test


reformat:
	go fmt ./...
	go vet ./...
	goimports -w $(shell find . -type f -name '*.go' -not -path "./vendor/*")


test:
	go test -v ./... 


cover:	
	@go test -coverprofile=coverage.out ./... &&  go tool cover -func=coverage.out	
	@rm -f coverage.out
