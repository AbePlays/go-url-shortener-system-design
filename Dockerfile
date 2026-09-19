FROM golang:1.27-alpine

WORKDIR /app

COPY . .

RUN go build -o server .

ENV BASE_URL=http://localhost:8080
ENV PORT=8080
EXPOSE 8080

CMD ["./server"]
