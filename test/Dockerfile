# Use the official Go image as the base image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /test

# Copy go.mod and go.sum files first for dependency resolution
COPY go.mod ./

# Download dependencies (this will be cached if unchanged)
RUN go mod download

# Copy the remaining source code
COPY . .

# Build the Go application
RUN go build -o myapp .

# Set the command to run the executable
CMD ["./myapp"]
