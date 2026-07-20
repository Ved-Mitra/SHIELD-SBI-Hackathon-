FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -o risk-url-engine ./cmd/risk-url-engine

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/risk-url-engine .

EXPOSE 8083

CMD ["./risk-url-engine"]