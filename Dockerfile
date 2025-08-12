# Use the official Golang image from the Docker Hub
FROM golang:1.24.3-alpine as builder

# Set the current working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies (will be cached if unchanged)
RUN go mod tidy

# Copy the source code into the container
COPY . .

# Set the working directory to where the main.go file is located
WORKDIR /app/cmd/server

# Build the Go app
RUN go build -o main .

# Stage 2: Build a smaller container to run the app
FROM alpine:latest

# Install CA certificates for https connections
RUN apk --no-cache add ca-certificates

# Create /app directory for the app
WORKDIR /app

# Copy the built executable from the builder stage
COPY --from=builder /app/cmd/server/main .

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
