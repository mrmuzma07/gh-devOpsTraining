# Multi-stage Dockerfile untuk efisiensi ukuran image
# Stage 1: Build binary Go
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o go-app main.go

# Stage 2: Minimal runtime image (Alpine Linux ~10MB)
FROM alpine:3.19
RUN apk add --no-cache curl
WORKDIR /app
COPY --from=builder /src/go-app .
EXPOSE 8080
CMD ["./go-app"]
