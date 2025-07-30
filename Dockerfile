# ixdx/esp-power-on:2025_07_30

# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o esp-power-on

# Stage 2: Minimal runtime image
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/esp-power-on .

EXPOSE 8080

ENTRYPOINT ["./esp-power-on"]
