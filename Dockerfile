FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN (cd cmd/shortener && go build -buildvcs=false -o shortener)

EXPOSE 8080

CMD ["./cmd/shortener/shortener"]
