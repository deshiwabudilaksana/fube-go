# Build stage
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# Use CGO_ENABLED=0 for a static binary that works in alpine
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS/external API calls
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy the templates directory
COPY --from=builder /app/templates ./templates

# The application listens on the port specified by the PORT environment variable
# Railway and Koyeb will inject this variable and handle the port mapping
EXPOSE 8080

# Command to run the executable
ENTRYPOINT ["./main"]
