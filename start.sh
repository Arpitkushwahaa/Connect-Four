#!/bin/bash

echo "======================================"
echo " 4 in a Row - Connect Four Game"
echo " Starting all services..."
echo "======================================"
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "ERROR: Docker is not installed!"
    echo "Please install Docker and Docker Compose first."
    exit 1
fi

echo "Docker is installed!"
echo ""

# Start services
echo "Starting services with Docker Compose..."
docker-compose up -d

if [ $? -ne 0 ]; then
    echo ""
    echo "ERROR: Failed to start services!"
    echo "Check the error messages above."
    exit 1
fi

echo ""
echo "======================================"
echo " Services are starting..."
echo " Please wait 30-60 seconds for all"
echo " services to be fully ready."
echo "======================================"
echo ""
echo " Frontend:  http://localhost:3000"
echo " Backend:   http://localhost:8080"
echo " WebSocket: ws://localhost:8080/ws"
echo ""
echo "To view logs: docker-compose logs -f"
echo "To stop:      docker-compose down"
echo ""
