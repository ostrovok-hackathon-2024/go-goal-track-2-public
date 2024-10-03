FROM debian:latest

# Install Go and required dependencies
RUN apt-get update && apt-get install -y \
    golang-1.19 \
    libc6 \
    && rm -rf /var/lib/apt/lists/*


# Set Go environment variables
ENV PATH="/usr/lib/go-1.19/bin:${PATH}"
ENV GOPATH="/go"
ENV PATH="${GOPATH}/bin:${PATH}"

# Set destination for COPY
WORKDIR /app

# Copy artifacts directory
COPY artifacts ./artifacts

COPY go/lib/ /usr/local/lib
COPY go/lib/ /root/local/lib

# Copy Go files
COPY go ./go

# Set working directory to go folder
WORKDIR /app/go

# Download Go modules
RUN go mod download

# Build
RUN GOOS=linux CGO_ENABLED=1 go build -o /usr/local/bin/cli ./cmd/cli/main.go

# Run
ENTRYPOINT ["cli"]
