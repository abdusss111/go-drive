#!/bin/bash

# GoDrive Deployment Script for GCP VM
# Run this after setup-gcp-vm.sh

set -e

echo "🚀 Deploying GoDrive application..."

# Check if .env exists
if [ ! -f .env ]; then
    echo "❌ Error: .env file not found!"
    echo "Please copy .env.template to .env and configure it"
    exit 1
fi

# Pull latest code
echo "📥 Pulling latest code..."
git pull origin main || echo "⚠️  Not a git repository or no remote configured"

# Stop existing containers
echo "🛑 Stopping existing containers..."
docker-compose down || true

# Build and start services
echo "🏗️  Building and starting services..."
docker-compose up -d --build

# Wait for database to be ready
echo "⏳ Waiting for database to be ready..."
sleep 10

# Run database migrations
echo "🗄️  Running database migrations..."
docker-compose exec -T api migrate -path /app/migrations -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=${POSTGRES_SSL_MODE}" up || echo "⚠️  Migrations may have already been applied"

# Configure Nginx (if not already done)
if [ -f "scripts/setup-nginx.sh" ] && [ ! -f "/etc/nginx/sites-enabled/godrive-api" ]; then
    echo "🌐 Setting up Nginx reverse proxy..."
    ./scripts/setup-nginx.sh
fi

# Show service status
echo ""
echo "✅ Deployment complete!"
echo ""
echo "📊 Service Status:"
docker-compose ps

echo ""
echo "🌐 Access your services:"
echo "  - API: http://$(curl -s ifconfig.me):8081"
echo "  - Swagger Docs: http://$(curl -s ifconfig.me):8081/swagger/index.html"
echo "  - Grafana: http://$(curl -s ifconfig.me):3001 (admin/change-me)"
echo "  - Prometheus: http://$(curl -s ifconfig.me):9091"
echo "  - MinIO Console: http://$(curl -s ifconfig.me):9003"
echo ""
echo "📝 View logs: docker-compose logs -f"
echo "🔄 Restart: docker-compose restart"
echo "🛑 Stop: docker-compose down"
