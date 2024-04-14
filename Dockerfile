FROM golang:1.22-alpine

WORKDIR /app

COPY cmd/banner-service/banner_service.go /app/
COPY internal/ /app/internal/
COPY go.mod /app/
COPY go.sum /app/

RUN go mod download

RUN go build -o banner-service /app/banner_service.go

EXPOSE 8080

CMD ["/app/banner-service"]