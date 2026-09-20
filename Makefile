.PHONY: test test-db build run dev fmt fmt-check vet clean docker-build docker-run compose-up compose-down compose-reset compose-logs

test:
	go test ./... -v

test-db:
	DATABASE_URL="postgres://postgres:postgres@localhost:5432/urlshortener?sslmode=disable" go test ./... -v

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

compose-up:
	docker compose up --build

compose-down:
	docker compose down

compose-reset:
	docker compose down -v

compose-logs:
	docker compose logs -f
