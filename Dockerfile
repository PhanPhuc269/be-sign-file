# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Copy .env file
COPY .env .env

# Build the Go app
RUN go build -o main .

# Run stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .
# Copy .env from builder
COPY --from=builder /app/.env .env
# Copy logs.html from builder
COPY --from=builder /app/logs.html logs.html
# Copy user_management_frontend.html from builder
COPY --from=builder /app/user_management_frontend.html user_management_frontend.html
# Copy all email templates from builder
COPY --from=builder /app/utils/email-template/ ./utils/email-template/
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./main", "--migrate", "--seed", "--run"]