# Use the official Golang image as the base
FROM golang:1.21 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files first for caching
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY src/ ./src/

# Set the working directory to where main.go is located
WORKDIR /app/src

# Compile the Go application as a static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/gateway

# Use a minimal base image for the final container
FROM alpine:latest

# Set the working directory for the final container
WORKDIR /app

# Install necessary dependencies
RUN apk --no-cache add ca-certificates

# Copy the compiled binary from the builder stage
COPY --from=builder /app/gateway .

# Ensure the binary has execution permissions
RUN chmod +x /app/gateway

# Set the entrypoint
ENTRYPOINT ["/app/gateway"]
