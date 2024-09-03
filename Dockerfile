# Use the official Go image as the base image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /labCode

# Copy the source code into the container
COPY . .

# Build the Go application

# Set the command to run the executable
