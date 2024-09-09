# Use an official Golang image as the base image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod  ./
RUN go mod download

# Copy the entire source code
COPY . .

# Build the Go application
RUN go build -o kademlia_app ./cmd/main.go

# Expose the port your application will run on
EXPOSE 8000

# Run the application
CMD ["./kademlia_app"]
