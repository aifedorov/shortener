# Stage 1: Build
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN cd cmd/shortener && go build -buildvcs=false -o shortener

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/cmd/shortener/shortener /app/shortener

COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080

CMD ["./shortener"]
