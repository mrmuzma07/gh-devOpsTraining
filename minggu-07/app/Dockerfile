# Multi-stage build untuk Go App dengan OpenTelemetry
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod/go.sum dulu untuk caching layer
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY main.go ./

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server .

# Runtime image
FROM alpine:3.18
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/server .

# Variabel env untuk OTel — bisa di-override di Kubernetes
ENV OTEL_EXPORTER_OTLP_ENDPOINT="grafana-alloy-otel.mini-prod.svc.cluster.local:4317"
ENV OTEL_SERVICE_NAME="go-app"

EXPOSE 8080
CMD ["./server"]