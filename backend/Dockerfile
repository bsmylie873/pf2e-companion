# ── Build stage ──
FROM golang:1.24-alpine AS builder

WORKDIR /build
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# ── Runtime stage ──
FROM alpine:3.21

RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /build/server .

EXPOSE 8080
CMD ["./server"]
