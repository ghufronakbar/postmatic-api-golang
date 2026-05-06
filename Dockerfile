# Tahap 1: Build binary
FROM golang:1.25.4-alpine AS builder

WORKDIR /app

# Copy file dependency dan download module
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh source code
COPY . .

# Build aplikasi dari entry point yang Anda sebutkan
RUN CGO_ENABLED=0 GOOS=linux go build -o postmatic-api ./cmd/api/main.go

# Tahap 2: Setup image super ringan untuk runtime
FROM alpine:latest

WORKDIR /app

# Copy timezone data (penting jika aplikasi bergantung pada waktu)
RUN apk add --no-cache tzdata

# Copy binary dari tahap builder
COPY --from=builder /app/postmatic-api .

# Expose port sesuai dengan env
EXPOSE 2001

# Jalankan aplikasi
CMD ["./postmatic-api"]