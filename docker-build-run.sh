#!/bin/bash

IMAGE_NAME="literary_lions"
CONTAINER_NAME="literary_lions"

# Build Docker image
echo "Building Docker image..."
docker build -t $IMAGE_NAME .

# Check if the Docker image was built successfully
if [ $? -eq 0 ]; then
    echo "Docker image built successfully: $IMAGE_NAME"
else
    echo "Failed to build Docker image: $IMAGE_NAME"
    exit 1
fi

# Check if a container with the same name is already running
if [ "$(docker ps -aq -f name=$CONTAINER_NAME)" ]; then
    echo "Stopping and removing existing container: $CONTAINER_NAME"
    docker stop $CONTAINER_NAME && docker rm $CONTAINER_NAME
fi

# Run Docker container
echo "Running Docker container..."
docker run --name $CONTAINER_NAME -d -p 3000:3000 $IMAGE_NAME

# Check if the container is running
if [ "$(docker ps -q -f name=$CONTAINER_NAME)" ]; then
    echo "Docker container is running: $CONTAINER_NAME"
else
    echo "Failed to run Docker container: $CONTAINER_NAME"
    exit 1
fi