# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Statically compile the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gate3 ./cmd/gate3

# Final stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/gate3 .

# Make sure the certs directory can be mounted
RUN mkdir -p /root/certs/gate2

EXPOSE 8082
CMD ["./gate3"]
