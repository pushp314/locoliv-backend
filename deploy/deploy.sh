#!/bin/bash
# LocoLive Backend - Full Deployment Script
# Run this AFTER setup-ec2.sh and after copying files to /opt/locolive

set -e

APP_DIR="/opt/locolive"
cd $APP_DIR

echo "🚀 Deploying LocoLive Backend"
echo "=============================="

# Check .env exists
if [ ! -f ".env" ]; then
    echo "❌ ERROR: .env file not found!"
    echo "   Copy deploy/.env.production to .env and fill in values"
    exit 1
fi

# Stop existing containers
echo "🛑 Stopping existing containers..."
docker-compose -f deploy/docker-compose.prod.yml down || true

# Pull latest images
echo "📦 Pulling latest images..."
docker-compose -f deploy/docker-compose.prod.yml pull

# Build API
echo "🔨 Building API container..."
docker-compose -f deploy/docker-compose.prod.yml build api

# Start database and redis first
echo "🗄️ Starting database and Redis..."
docker-compose -f deploy/docker-compose.prod.yml up -d postgres redis
sleep 10

# Run migrations
echo "📊 Running database migrations..."
docker-compose -f deploy/docker-compose.prod.yml --profile migrate run --rm migrate

# Start API
echo "🚀 Starting API..."
docker-compose -f deploy/docker-compose.prod.yml up -d api

# Setup systemd service
echo "⚙️ Configuring systemd service..."
sudo cp deploy/locolive.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable locolive

# Configure Nginx
echo "🌐 Configuring Nginx..."
sudo cp deploy/nginx.conf /etc/nginx/sites-available/locolive
sudo ln -sf /etc/nginx/sites-available/locolive /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t && sudo systemctl reload nginx

# Check health
echo "🏥 Checking API health..."
sleep 5
curl -s http://localhost:8080/health || echo "Warning: Health check failed"

echo ""
echo "✅ Deployment complete!"
echo ""
echo "🔗 Your API is running at: http://$(curl -s ifconfig.me):80"
echo ""
echo "📋 Useful commands:"
echo "   View logs:     docker-compose -f deploy/docker-compose.prod.yml logs -f"
echo "   Restart:       sudo systemctl restart locolive"
echo "   Status:        docker-compose -f deploy/docker-compose.prod.yml ps"
echo ""
