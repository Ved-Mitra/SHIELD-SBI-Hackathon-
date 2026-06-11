# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o gate2 ./cmd/gate2

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/gate2 .

# Copy certificates (standard dev certs)
COPY certs/ ./certs/

# Expose port
EXPOSE 8080

# Run the application
CMD ["./gate2"]
