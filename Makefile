.PHONY: test build run dev fmt fmt-check vet clean docker-build docker-run

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

fmt-check:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "The following files are not gofmt'd:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

vet:
	go vet ./...

clean:
	rm -rf bin/ tmp/

docker-build:
	docker build -t url-shortener .

docker-run:
	docker run -p 8080:8080 url-shortener
