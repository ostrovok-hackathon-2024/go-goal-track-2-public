FROM alpine:latest

RUN apk add --no-cache gcc go=1.22.7-r0 libc6-compat musl-dev

RUN apk --no-cache add ca-certificates

ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/go"
ENV PATH="${GOPATH}/bin:${PATH}"

WORKDIR /app

COPY artifacts ./artifacts

COPY go/lib/libcatboostmodel.so /usr/local/lib/libcatboostmodel.so

COPY go ./go

WORKDIR /app/go

RUN go mod download

ENV PATH="/usr/local/lib:${PATH}"

RUN GOOS=linux CGO_ENABLED=1 go build -o /usr/local/bin/cli ./cmd/cli/main.go

ENTRYPOINT ["cli"]
