# Use an official Golang image as the base image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy Go module files and download dependencies
COPY go.mod ./

# Copy the entire source code
COPY . .

# Copy the entrypoint script into the container
COPY scripts/entrypoint.sh scripts/entrypoint.sh

# Give the entrypoint script execution permission
RUN chmod +x scripts/entrypoint.sh

# Build the Go application
RUN go build -o kademlia_app ./cmd/main.go

# Expose the port your application will run on
EXPOSE 8080

# Set the entrypoint to the entrypoint script
ENTRYPOINT ["./scripts/entrypoint.sh"]

# Run the Go application (this gets executed after the entrypoint script)
CMD ["./kademlia_app"]
