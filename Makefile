.PHONY: test build run dev fmt vet clean docker-build docker-run

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

docker-build:
	docker build -t url-shortener .

docker-run:
	docker run -p 8080:8080 url-shortener
