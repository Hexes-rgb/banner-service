FROM golang:1.22-alpine

WORKDIR /app

COPY src/ ./

RUN go mod download

RUN go build -o main .

EXPOSE 8080

CMD ["/app/main"]