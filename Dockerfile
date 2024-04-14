FROM golang:1.22-alpine

WORKDIR /app

COPY cmd/banner-service/ ./

COPY go.mod ./

COPY go.sum ./

RUN go mod download

RUN go build -o banner-service .

EXPOSE 8080

CMD ["/app/banner-service"]