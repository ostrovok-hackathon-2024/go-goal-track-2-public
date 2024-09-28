# FROM golang:1.23 AS builder
# WORKDIR /app
# COPY . .
# RUN CGO_ENABLED=0 GOOS=linux go build -o tagger cmd/taggmain.go

# FROM alpine:latest
# WORKDIR /app
# COPY --from=builder /app/tagger .
# ENTRYPOINT ["./tagger"]