#!/bin/sh

# Get the dynamically assigned IP address of the container
CONTAINER_IP=$(hostname -i)

# Export the IP address as an environment variable
export CONTAINER_IP

echo "Container IP Address: $CONTAINER_IP"

# Call the main process (Go app) with the new environment variable
exec "$@"
