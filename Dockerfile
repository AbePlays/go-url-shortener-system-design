FROM golang:1.27-alpine AS builder

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

WORKDIR /app

COPY . .

RUN go build -o server .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/server .
COPY db/migrations ./db/migrations

ENV BASE_URL=http://localhost:8080
ENV PORT=8080
EXPOSE 8080

CMD ["sh", "-c", "migrate -path db/migrations -database \"$DATABASE_URL\" up && exec ./server"]
