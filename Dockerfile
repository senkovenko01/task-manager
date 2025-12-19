FROM golang:1.25-alpine AS builder

# Install build dependencies for sqlite3
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application
# CGO_ENABLED=1 is required for sqlite3
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o task-manager .

FROM alpine:latest

# Install runtime dependencies for sqlite3 and wget for healthcheck
RUN apk --no-cache add ca-certificates sqlite wget

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/task-manager .

# Copy database file if it exists
VOLUME ["/app/data"]

EXPOSE 8080

CMD ["./task-manager"]
