#!/bin/bash
set -e

echo "Starting URL Shortener setup..."

if [ ! -f .env ]; then
    echo "Creating .env file from .env.example..."
    cp .env.example .env
    echo ".env file created. Please review and update the configuration if needed."
fi

echo "Starting Docker Compose stack..."
docker compose up --build -d

echo "Waiting for PostgreSQL to be ready..."
sleep 5

echo "Checking service health..."
max_attempts=30
attempt=0

while [ $attempt -lt $max_attempts ]; do
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        echo "✅ Service is healthy!"
        break
    fi
    
    attempt=$((attempt + 1))
    echo "Waiting for service to be ready... ($attempt/$max_attempts)"
    sleep 2
done

if [ $attempt -eq $max_attempts ]; then
    echo "❌ Service failed to start. Check logs with: docker compose logs app"
    exit 1
fi

echo ""
echo "🚀 URL Shortener is running!"
echo ""
echo "API Endpoint: http://localhost:8080"
echo "Health Check: http://localhost:8080/health"
echo ""
echo "Example commands:"
echo "  Create short URL:"
echo "    curl -X POST http://localhost:8080/api/urls -H 'Content-Type: application/json' -d '{\"url\": \"https://github.com\"}'"
echo ""
echo "  List URLs:"
echo "    curl http://localhost:8080/api/urls"
echo ""
echo "To stop: docker compose down"
echo "To view logs: docker compose logs -f app"
