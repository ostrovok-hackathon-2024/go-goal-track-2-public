FROM golang:1.23.1 AS builder

WORKDIR /build

COPY --from=root . .
RUN go mod download

RUN go build -ldflags="-s -w" -o main ./cmd/api/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder ["/build/main", "/"]

CMD ["openssl req -config example-com.conf -new -x509 -sha256 -newkey rsa:2048 -nodes \
    -keyout example-com.key.pem -days 365 -out example-com.cert.pem"]

ENTRYPOINT ["./main"]
