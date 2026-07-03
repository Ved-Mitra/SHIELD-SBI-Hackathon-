FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Must enable CGO for confluent-kafka-go
RUN CGO_ENABLED=1 GOOS=linux go build -o consumer ./cmd/consumer

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/consumer .

CMD ["./consumer"]
