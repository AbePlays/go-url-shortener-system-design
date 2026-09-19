.PHONY: test build run dev fmt vet clean

test:
	go test ./... -v

build:
	go build -o bin/server .

run:
	go run .

dev:
	air

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -rf bin/ tmp/
