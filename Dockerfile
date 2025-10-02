# Stage 1: Build the Go application
FROM golang:1.25-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to leverage Docker's caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
# CGO_ENABLED=0 is important for creating a statically linked binary
# -o specifies the output binary name
# -ldflags="-s -w" reduces binary size by stripping debug info
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o /app/main .

# Stage 2: Create a minimal production image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Install root CA certificates for outbound HTTPS (e.g., to Postgres, OAuth, etc.)
RUN apk --no-cache add ca-certificates

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Expose the port your application listens on (Cloud Run expects 8080 by default)
EXPOSE 8080

# Run Gin in release mode in production
ENV GIN_MODE=release

# Command to run the application
CMD ["./main"]