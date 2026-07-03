FROM golang:alpine AS builder
WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

# Minimal final image
FROM alpine:latest
WORKDIR /app

# Add timezone data
RUN apk --no-cache add tzdata

COPY --from=builder /app/server .
# Note: In production, .env should be mounted or injected via docker-compose/K8s
# but we copy it here as fallback if it exists
COPY .env* ./

EXPOSE 8080
CMD ["./server"]
