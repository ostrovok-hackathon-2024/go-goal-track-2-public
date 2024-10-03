FROM golang:1.23.1

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY ../artifacts ./artifacts

# Build
RUN GOOS=linux go build -o /usr/local/bin/api ./cmd/api/main.go


# Run
ENTRYPOINT ["api"]
