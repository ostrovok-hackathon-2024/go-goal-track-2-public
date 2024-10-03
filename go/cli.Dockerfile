FROM golang:1.23.1 AS builder

WORKDIR /build

COPY . .
RUN go mod download

RUN go build -ldflags="-s -w" -o main ./cmd/cli/main.go

FROM alpine:latest

COPY --from=builder ["/build/main", "/"]

ENTRYPOINT ["./main"]
