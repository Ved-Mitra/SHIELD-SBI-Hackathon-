FROM golang:1.22-alpine AS builder

# Install build dependencies for confluent-kafka-go (requires cgo)
RUN apk add --no-cache gcc musl-dev libc-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Must enable CGO for confluent-kafka-go
RUN CGO_ENABLED=1 GOOS=linux go build -o consumer ./cmd/consumer

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/consumer .

CMD ["./consumer"]
