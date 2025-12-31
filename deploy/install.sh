#!/bin/bash
# LocoLive One-Click Installer for Ubuntu EC2
# Usage: sudo ./install.sh

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}🚀 LocoLive Backend - Auto Installer${NC}"
echo "========================================"

if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Please run as root (use sudo)${NC}"
    exit 1
fi

# 1. Update System
echo -e "\n${BLUE}📦 Updating system packages...${NC}"
apt update && apt upgrade -y
apt install -y curl wget git unzip build-essential nginx redis-server postgresql postgresql-contrib certbot python3-certbot-nginx

# 2. Install Go 1.22
echo -e "\n${BLUE}🐹 Installing Go 1.22...${NC}"
if ! command -v go &> /dev/null; then
    wget -q https://go.dev/dl/go1.22.5.linux-amd64.tar.gz
    rm -rf /usr/local/go
    tar -C /usr/local -xzf go1.22.5.linux-amd64.tar.gz
    rm go1.22.5.linux-amd64.tar.gz
    
    # Add to path for this session
    export PATH=$PATH:/usr/local/go/bin
    
    # Add to bashrc if not present
    if ! grep -q "/usr/local/go/bin" /home/ubuntu/.bashrc; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> /home/ubuntu/.bashrc
        echo 'export PATH=$PATH:/home/ubuntu/go/bin' >> /home/ubuntu/.bashrc
    fi
else
    echo "Go already installed."
fi

# 3. Install Migrate Tool
echo -e "\n${BLUE}🔄 Installing Database Migration tool...${NC}"
if ! command -v migrate &> /dev/null; then
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
    mv migrate /usr/local/bin/migrate
    chmod +x /usr/local/bin/migrate
else
    echo "Migrate tool already installed."
fi

# 4. Generate Secrets
echo -e "\n${BLUE}🔐 Generating secure passwords...${NC}"
DB_PASSWORD=$(openssl rand -base64 32 | tr -d '/+=' | head -c 24)
JWT_SECRET=$(openssl rand -base64 64 | tr -d '\n')

echo "Generated Database Password: $DB_PASSWORD"
echo "Generated JWT Secret"

# 5. Configure PostgreSQL
echo -e "\n${BLUE}🗄️  Configuring PostgreSQL...${NC}"
systemctl start postgresql
systemctl enable postgresql

# Check if user exists, if not create
if ! sudo -u postgres psql -tAc "SELECT 1 FROM pg_roles WHERE rolname='locolive'" | grep -q 1; then
    sudo -u postgres psql -c "CREATE USER locolive WITH PASSWORD '$DB_PASSWORD';"
    sudo -u postgres psql -c "CREATE DATABASE locolive;"
    sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE locolive TO locolive;"
    sudo -u postgres psql -d locolive -c "GRANT ALL ON SCHEMA public TO locolive;"
    echo "Database user and DB created."
else
    # Update password if user exists
    sudo -u postgres psql -c "ALTER USER locolive WITH PASSWORD '$DB_PASSWORD';"
    echo "Updated existing database user password."
fi

# 6. Setup Directory & Repo
echo -e "\n${BLUE}📂 Setting up application directory...${NC}"
APP_DIR="/opt/locolive"
mkdir -p $APP_DIR

# If current directory contains go.mod, copy files, else clone (assuming script run from cloned repo)
if [ -f "../go.mod" ]; then
    echo "Copying files from parent directory..."
    cp -r .. $APP_DIR/src
    mv $APP_DIR/src/* $APP_DIR/
    rm -r $APP_DIR/src
    chown -R ubuntu:ubuntu $APP_DIR
    cd $APP_DIR
else
    echo -e "${RED}Please run this script from inside the deploy/ folder of the cloned repository.${NC}"
    exit 1
fi

# 7. Configure Environment
echo -e "\n${BLUE}📝 Configuring .env file...${NC}"

# Ask for Google Credentials
echo -e "${GREEN}Enter Google Client ID (leave empty to set later):${NC}"
read GOOGLE_CLIENT_ID
echo -e "${GREEN}Enter Google Client Secret (leave empty to set later):${NC}"
read GOOGLE_CLIENT_SECRET

GOOGLE_CLIENT_ID=${GOOGLE_CLIENT_ID:-"change_me"}
GOOGLE_CLIENT_SECRET=${GOOGLE_CLIENT_SECRET:-"change_me"}

cat > $APP_DIR/.env <<EOL
PORT=8080
ENV=production

DATABASE_URL=postgres://locolive:${DB_PASSWORD}@localhost:5432/locolive?sslmode=disable
DB_HOST=localhost
DB_PORT=5432
DB_USER=locolive
DB_PASSWORD=${DB_PASSWORD}
DB_NAME=locolive

REDIS_URL=redis://localhost:6379

JWT_SECRET=${JWT_SECRET}
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

GOOGLE_CLIENT_ID=${GOOGLE_CLIENT_ID}
GOOGLE_CLIENT_SECRET=${GOOGLE_CLIENT_SECRET}

LOG_LEVEL=info
EOL

chown ubuntu:ubuntu $APP_DIR/.env
chmod 600 $APP_DIR/.env

# 8. Run Migrations
echo -e "\n${BLUE}📊 Running migrations...${NC}"
sudo -u ubuntu migrate -database "postgres://locolive:${DB_PASSWORD}@localhost:5432/locolive?sslmode=disable" -path $APP_DIR/db/migrations up

# 9. Build App
echo -e "\n${BLUE}🔨 Building application...${NC}"
cd $APP_DIR
export PATH=$PATH:/usr/local/go/bin
go mod download
go build -ldflags="-w -s" -o bin/api ./cmd/api
chown -R ubuntu:ubuntu bin/api

# 10. Setup Systemd
echo -e "\n${BLUE}⚙️  Configuring systemd service...${NC}"
cat > /etc/systemd/system/locolive.service <<EOL
[Unit]
Description=LocoLive Backend API
After=network.target postgresql.service redis-server.service

[Service]
Type=simple
User=ubuntu
Group=ubuntu
WorkingDirectory=$APP_DIR
ExecStart=$APP_DIR/bin/api
Restart=always
RestartSec=5
EnvironmentFile=$APP_DIR/.env
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOL

systemctl daemon-reload
systemctl enable locolive
systemctl restart locolive

# 11. Setup Nginx
echo -e "\n${BLUE}🌐 Configuring Nginx...${NC}"
cat > /etc/nginx/sites-available/locolive <<EOL
server {
    listen 80;
    server_name launchit.co.in www.launchit.co.in;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    location /health {
        proxy_pass http://127.0.0.1:8080/health;
        access_log off;
    }
}
EOL

ln -sf /etc/nginx/sites-available/locolive /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default
nginx -t
systemctl reload nginx

echo -e "${GREEN}✅ Installation Complete!${NC}"
echo "-----------------------------------"
echo "API is running at http://launchit.co.in"
echo ""
echo "Only one step left:"
echo "Run: sudo certbot --nginx -d launchit.co.in"
echo "to enable HTTPS."
