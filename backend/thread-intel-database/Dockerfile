FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Pure Go build, no CGO required for segmentio/kafka-go
RUN CGO_ENABLED=0 GOOS=linux go build -o consumer ./cmd/consumer

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/consumer .

CMD ["./consumer"]
