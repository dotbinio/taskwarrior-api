# Build stage
FROM golang:1.25.4-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Install swag for Swagger docs generation
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/

# Generate Swagger docs and build
RUN /go/bin/swag init -g cmd/server/main.go -o docs
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o taskwarrior-api ./cmd/server

# Runtime stage
FROM alpine:latest

# Install runtime dependencies and Taskwarrior
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    wget \
    task

# Set working directory
WORKDIR /app

# Copy API binary from Go builder
COPY --from=builder /build/taskwarrior-api .

# Create directory for Taskwarrior data and initialize config
RUN mkdir -p /root/.task && \
    echo "data.location=/root/.task" > /root/.taskrc && \
    echo "confirmation=no" >> /root/.taskrc

# Verify Taskwarrior installation
RUN task --version

# Expose port
EXPOSE 8080

CMD ["./taskwarrior-api"]
