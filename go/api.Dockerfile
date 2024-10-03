FROM alpine:latest

# Install build dependencies
RUN apk add --no-cache gcc libc6-compat musl-dev

# Install Go 1.23.1
RUN apk add --no-cache go=1.23.1-r0 --repository=http://dl-cdn.alpinelinux.org/alpine/edge/community

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

RUN GOOS=linux CGO_ENABLED=1 go build -o /usr/local/bin/api ./cmd/api/main.go

ENTRYPOINT ["api"]
