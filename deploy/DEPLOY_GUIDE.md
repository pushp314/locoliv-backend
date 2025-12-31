# 🚀 Complete AWS EC2 Deployment Guide
## LocoLive Backend → launchit.co.in

**Your Setup:**
- Domain: `launchit.co.in`
- PEM Key: `locolive.pem`
- Instance: Ubuntu 22.04 on t3.micro

---

## 📋 TABLE OF CONTENTS

1. [Create EC2 Instance](#step-1-create-ec2-instance)
2. [Configure Security Group](#step-2-configure-security-group)
3. [Point Domain to EC2](#step-3-point-domain-to-ec2)
4. [SSH into EC2](#step-4-ssh-into-ec2)
5. [Update System Packages](#step-5-update-system-packages)
6. [Install Docker](#step-6-install-docker)
7. [Install Docker Compose](#step-7-install-docker-compose)
8. [Install Nginx](#step-8-install-nginx)
9. [Install Git](#step-9-install-git)
10. [Clone Repository](#step-10-clone-repository)
11. [Configure Environment](#step-11-configure-environment)
12. [Start Database & Redis](#step-12-start-database--redis)
13. [Run Migrations](#step-13-run-migrations)
14. [Start the API](#step-14-start-the-api)
15. [Configure Nginx](#step-15-configure-nginx)
16. [Install SSL Certificate](#step-16-install-ssl-certificate)
17. [Setup Auto-Start](#step-17-setup-auto-start)
18. [Configure GitHub CI/CD](#step-18-configure-github-cicd)
19. [Verify Deployment](#step-19-verify-deployment)

---

## STEP 1: Create EC2 Instance

### In AWS Console (https://console.aws.amazon.com)

1. Go to **EC2 → Instances → Launch instances**

2. Configure:
   ```
   Name:           locolive-backend
   AMI:            Ubuntu Server 22.04 LTS (HVM), SSD Volume Type
   Architecture:   64-bit (Arm) - for t3.micro
   Instance type:  t3.micro
   Key pair:       locolive (use your existing locolive.pem)
   ```

3. **Network settings** → Click "Edit"
   - Auto-assign public IP: **Enable**

4. **Configure storage**: 20 GB gp3 (default is fine)

5. Click **Launch instance**

6. **Note your EC2 Public IP** (e.g., `13.235.xxx.xxx`)

---

## STEP 2: Configure Security Group

### In AWS Console → EC2 → Security Groups

1. Find the security group attached to your instance
2. Click **Inbound rules → Edit inbound rules**
3. Add these rules:

| Type | Port Range | Source | Description |
|------|------------|--------|-------------|
| SSH | 22 | My IP | SSH access |
| HTTP | 80 | 0.0.0.0/0 | Web traffic |
| HTTPS | 443 | 0.0.0.0/0 | Secure web traffic |
| Custom TCP | 8080 | 0.0.0.0/0 | API direct (optional) |

4. Click **Save rules**

---

## STEP 3: Point Domain to EC2

### In your domain registrar (GoDaddy/Namecheap/etc.)

1. Go to DNS Management for `launchit.co.in`

2. Add/Update these DNS records:

| Type | Host | Value | TTL |
|------|------|-------|-----|
| A | @ | YOUR_EC2_PUBLIC_IP | 600 |
| A | www | YOUR_EC2_PUBLIC_IP | 600 |

3. Wait 5-10 minutes for DNS propagation

4. Verify:
   ```bash
   # On your Mac
   nslookup launchit.co.in
   # Should show your EC2 IP
   ```

---

## STEP 4: SSH into EC2

```bash
# On your Mac - Terminal

# Navigate to where your key is stored
cd ~/Downloads  # or wherever locolive.pem is

# Set correct permissions
chmod 400 locolive.pem

# Connect to EC2
ssh -i locolive.pem ubuntu@YOUR_EC2_PUBLIC_IP

# Example:
# ssh -i locolive.pem ubuntu@13.235.123.456
```

✅ You should see: `ubuntu@ip-xxx-xxx-xxx-xxx:~$`

---

## STEP 5: Update System Packages

```bash
# Run these commands ON EC2

# Update package list
sudo apt update

# Upgrade all packages
sudo apt upgrade -y

# Install essential tools
sudo apt install -y curl wget git unzip
```

---

## STEP 6: Install Docker

```bash
# Remove old Docker versions (if any)
sudo apt remove docker docker-engine docker.io containerd runc 2>/dev/null || true

# Install Docker using official script
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add your user to docker group (no sudo needed for docker commands)
sudo usermod -aG docker ubuntu

# Start Docker and enable on boot
sudo systemctl start docker
sudo systemctl enable docker

# IMPORTANT: Apply group changes
newgrp docker

# Verify Docker installation
docker --version
# Expected: Docker version 24.x.x or higher

docker run hello-world
# Expected: "Hello from Docker!" message
```

---

## STEP 7: Install Docker Compose

```bash
# Download latest Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose

# Make it executable
sudo chmod +x /usr/local/bin/docker-compose

# Verify installation
docker-compose --version
# Expected: Docker Compose version v2.x.x
```

---

## STEP 8: Install Nginx

```bash
# Install Nginx
sudo apt install -y nginx

# Start and enable Nginx
sudo systemctl start nginx
sudo systemctl enable nginx

# Verify it's running
sudo systemctl status nginx
# Should show: active (running)

# Test in browser (temporary)
# Visit: http://YOUR_EC2_IP or http://launchit.co.in
# Should show: "Welcome to nginx!"
```

---

## STEP 9: Install Git

```bash
# Git is usually pre-installed, but let's make sure
sudo apt install -y git

# Verify
git --version
# Expected: git version 2.x.x

# Configure Git (optional but recommended)
git config --global user.name "Your Name"
git config --global user.email "your@email.com"
```

---

## STEP 10: Clone Repository

### First, push your code to GitHub (on your Mac):

```bash
# On your Mac
cd /Users/pushp314/Desktop/LocoLiv/locolive-backend

# Initialize git if not already
git init

# Add all files
git add .

# Commit
git commit -m "Initial commit - LocoLive Backend"

# Create GitHub repo at https://github.com/new
# Then add remote and push:
git remote add origin https://github.com/YOUR_USERNAME/locolive-backend.git
git branch -M main
git push -u origin main
```

### Now clone on EC2:

```bash
# On EC2

# Create app directory
sudo mkdir -p /opt/locolive
sudo chown ubuntu:ubuntu /opt/locolive

# Clone your repository
cd /opt/locolive
git clone https://github.com/YOUR_USERNAME/locolive-backend.git .

# Verify files are there
ls -la
# Should see: cmd/, internal/, deploy/, docker-compose.yml, etc.
```

---

## STEP 11: Configure Environment

```bash
# On EC2
cd /opt/locolive

# Copy production environment template
cp deploy/.env.production .env

# Generate a secure JWT secret (copy the output)
openssl rand -base64 64

# Edit the environment file
nano .env
```

### Edit `.env` with these values:

```env
# Server
PORT=8080
ENV=production

# Database - CHANGE THE PASSWORD!
DATABASE_URL=postgres://locolive:YOUR_SECURE_PASSWORD_HERE@postgres:5432/locolive?sslmode=disable
DB_HOST=postgres
DB_PORT=5432
DB_USER=locolive
DB_PASSWORD=YOUR_SECURE_PASSWORD_HERE
DB_NAME=locolive

# Redis
REDIS_URL=redis://redis:6379

# JWT - PASTE YOUR GENERATED SECRET HERE!
JWT_SECRET=PASTE_THE_64_CHAR_SECRET_FROM_OPENSSL_HERE
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# Google OAuth - Use your real credentials
GOOGLE_CLIENT_ID=your-google-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-google-client-secret

# Logging
LOG_LEVEL=info
```

**Save: `Ctrl+O`, `Enter`, `Ctrl+X`**

---

## STEP 12: Start Database & Redis

```bash
# On EC2
cd /opt/locolive

# Start PostgreSQL and Redis
docker-compose -f deploy/docker-compose.prod.yml up -d postgres redis

# Wait for them to be healthy (10-15 seconds)
sleep 15

# Verify containers are running
docker-compose -f deploy/docker-compose.prod.yml ps

# Expected output:
# NAME                          STATUS
# locolive-postgres-1          Up (healthy)
# locolive-redis-1             Up

# Test PostgreSQL connection
docker exec locolive-backend-postgres-1 psql -U locolive -c "SELECT 1;"
# Expected: shows "1" in a table
```

---

## STEP 13: Run Migrations

```bash
# On EC2
cd /opt/locolive

# Run database migrations
docker-compose -f deploy/docker-compose.prod.yml --profile migrate run --rm migrate

# Expected output:
# 1/u init_schema (XXms)

# Verify tables were created
docker exec locolive-backend-postgres-1 psql -U locolive -c "\dt"
# Should show: users, sessions, refresh_tokens, etc.
```

---

## STEP 14: Start the API

```bash
# On EC2
cd /opt/locolive

# Build and start the API
docker-compose -f deploy/docker-compose.prod.yml up -d --build api

# Wait for it to start
sleep 10

# Check it's running
docker-compose -f deploy/docker-compose.prod.yml ps

# View logs
docker-compose -f deploy/docker-compose.prod.yml logs api

# Test the API locally
curl http://localhost:8080/health
# Expected: {"status":"ok","timestamp":"...","version":"1.0.0"}
```

---

## STEP 15: Configure Nginx

```bash
# On EC2

# Copy Nginx configuration
sudo cp /opt/locolive/deploy/nginx.conf /etc/nginx/sites-available/locolive

# Enable the site
sudo ln -sf /etc/nginx/sites-available/locolive /etc/nginx/sites-enabled/

# Remove default site
sudo rm -f /etc/nginx/sites-enabled/default

# Test Nginx configuration
sudo nginx -t
# Expected: "syntax is ok" and "test is successful"

# Reload Nginx
sudo systemctl reload nginx

# Test through Nginx
curl http://localhost/health
# Expected: {"status":"ok"...}

# Test from your domain
curl http://launchit.co.in/health
# Expected: {"status":"ok"...}
```

---

## STEP 16: Install SSL Certificate

```bash
# On EC2

# Install Certbot
sudo apt install -y certbot python3-certbot-nginx

# Get SSL certificate for your domain
sudo certbot --nginx -d launchit.co.in -d www.launchit.co.in

# Follow the prompts:
# 1. Enter your email
# 2. Agree to terms (A)
# 3. Share email? (N)
# 4. Redirect HTTP to HTTPS? (2 - Redirect)

# Verify SSL auto-renewal
sudo certbot renew --dry-run
# Should show: "Congratulations, all simulated renewals succeeded"

# Test HTTPS
curl https://launchit.co.in/health
# Expected: {"status":"ok"...}
```

---

## STEP 17: Setup Auto-Start

```bash
# On EC2

# Copy systemd service file
sudo cp /opt/locolive/deploy/locolive.service /etc/systemd/system/

# Reload systemd
sudo systemctl daemon-reload

# Enable auto-start on boot
sudo systemctl enable locolive

# Start the service
sudo systemctl start locolive

# Check status
sudo systemctl status locolive
# Should show: active (running)

# Verify after reboot (optional test)
# sudo reboot
# Then SSH back in and check: docker ps
```

---

## STEP 18: Configure GitHub CI/CD

### On GitHub (https://github.com/YOUR_USERNAME/locolive-backend)

1. Go to **Settings → Secrets and variables → Actions**

2. Click **New repository secret** and add:

| Secret Name | Value |
|-------------|-------|
| `EC2_HOST` | Your EC2 public IP (e.g., `13.235.123.456`) |
| `EC2_SSH_KEY` | Contents of your locolive.pem file |

### Get your PEM file content (on Mac):

```bash
# On your Mac
cat ~/Downloads/locolive.pem
# Copy the ENTIRE output including -----BEGIN RSA PRIVATE KEY----- and -----END RSA PRIVATE KEY-----
```

3. Paste the entire key content into the `EC2_SSH_KEY` secret

### Enable GitHub Actions:

1. Go to **Actions** tab in your repository
2. Click **Enable workflows** if prompted

### Now every push to `main` will auto-deploy! 🎉

---

## STEP 19: Verify Deployment

### Test all endpoints:

```bash
# Health check
curl https://launchit.co.in/health

# Register a user
curl -X POST https://launchit.co.in/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"SecurePass123","name":"Test User"}'

# Login
curl -X POST https://launchit.co.in/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"SecurePass123"}'
```

---

## 📋 USEFUL COMMANDS REFERENCE

```bash
# SSH into EC2
ssh -i ~/Downloads/locolive.pem ubuntu@YOUR_EC2_IP

# View all logs
docker-compose -f /opt/locolive/deploy/docker-compose.prod.yml logs -f

# View only API logs
docker-compose -f /opt/locolive/deploy/docker-compose.prod.yml logs -f api

# Restart API
docker-compose -f /opt/locolive/deploy/docker-compose.prod.yml restart api

# Restart everything
sudo systemctl restart locolive

# Stop everything
docker-compose -f /opt/locolive/deploy/docker-compose.prod.yml down

# Pull latest code and redeploy
cd /opt/locolive && git pull && docker-compose -f deploy/docker-compose.prod.yml up -d --build api

# Check disk space
df -h

# Check memory
free -h

# Check running containers
docker ps
```

---

## 🔧 TROUBLESHOOTING

### Container won't start
```bash
docker-compose -f deploy/docker-compose.prod.yml logs api
```

### Database connection error
```bash
docker-compose -f deploy/docker-compose.prod.yml logs postgres
docker exec locolive-backend-postgres-1 psql -U locolive -c "SELECT 1;"
```

### Nginx 502 Bad Gateway
```bash
# Check if API is running
curl http://localhost:8080/health

# Check Nginx error log
sudo tail -f /var/log/nginx/error.log
```

### SSL certificate issues
```bash
sudo certbot renew
sudo nginx -t
sudo systemctl reload nginx
```

### Port already in use
```bash
sudo lsof -i :8080
sudo kill -9 <PID>
```

---

## 💰 COST ESTIMATE

| Service | Monthly |
|---------|---------|
| EC2 t3.micro | ~$8.50 |
| EBS 20GB | ~$1.60 |
| Data Transfer | ~$2-5 |
| **Total** | **~$12-15** |

*First 12 months: 750 hrs/month free tier eligible*

---

## ✅ DEPLOYMENT CHECKLIST

- [ ] EC2 instance created
- [ ] Security group configured
- [ ] Domain DNS pointing to EC2
- [ ] Docker installed
- [ ] Docker Compose installed
- [ ] Nginx installed
- [ ] Git installed
- [ ] Repository cloned
- [ ] `.env` configured with secrets
- [ ] Database running
- [ ] Migrations executed
- [ ] API running
- [ ] Nginx configured
- [ ] SSL certificate installed
- [ ] Auto-start enabled
- [ ] GitHub secrets configured
- [ ] CI/CD tested

---

**🎉 Congratulations! Your LocoLive backend is now live at https://launchit.co.in**
