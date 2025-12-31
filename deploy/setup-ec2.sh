#!/bin/bash
# LocoLive Backend - EC2 Deployment Script
# Run this ON your EC2 Ubuntu instance

set -e

echo "🚀 LocoLive Backend Deployment Script"
echo "======================================"

# Update system
echo "📦 Updating system packages..."
sudo apt update && sudo apt upgrade -y

# Install Docker
echo "🐳 Installing Docker..."
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    sudo usermod -aG docker $USER
    rm get-docker.sh
    echo "Docker installed. You may need to log out and back in for group changes."
fi

# Install Docker Compose
echo "🐳 Installing Docker Compose..."
if ! command -v docker-compose &> /dev/null; then
    sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
fi

# Install Nginx
echo "🌐 Installing Nginx..."
sudo apt install -y nginx

# Create app directory
echo "📁 Setting up application directory..."
sudo mkdir -p /opt/locolive
sudo chown $USER:$USER /opt/locolive

echo ""
echo "✅ Base setup complete!"
echo ""
echo "Next steps:"
echo "1. Copy your project files to /opt/locolive"
echo "2. Create /opt/locolive/.env with production values"
echo "3. Run: cd /opt/locolive && docker-compose up -d"
echo "4. Configure Nginx (see nginx.conf)"
echo ""
