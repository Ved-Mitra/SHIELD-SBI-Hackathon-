# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Download dependencies first (layer cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a statically-linked binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-s -w" \
      -o gate1 ./cmd/gate1

# ── Stage 2: Minimal runtime image ────────────────────────────────────────────
FROM alpine:3.20

# ca-certificates is needed for TLS calls to Google/Apple APIs
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the binary
COPY --from=builder /app/gate1 .

# The certs directory is mounted at runtime (not baked into image).
# See docker-compose.yml for the volume mount.
# COPY certs/ ./certs/     ← don't bake secrets into the image

EXPOSE 8081

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8081/healthz || exit 1

ENTRYPOINT ["./gate1"]
