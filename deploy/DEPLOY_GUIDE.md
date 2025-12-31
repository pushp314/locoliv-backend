# 🚀 Complete AWS EC2 Deployment Guide (No Docker)
## LocoLive Backend → launchit.co.in

**Native Deployment using:**
- Go binary (compiled)
- PostgreSQL (native)
- Redis (native)
- Nginx (reverse proxy)
- systemd (process manager)
- Let's Encrypt SSL

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
6. [Install Go](#step-6-install-go)
7. [Install PostgreSQL](#step-7-install-postgresql)
8. [Install Redis](#step-8-install-redis)
9. [Install Nginx](#step-9-install-nginx)
10. [Install Migrate Tool](#step-10-install-migrate-tool)
11. [Clone Repository](#step-11-clone-repository)
12. [Configure Environment](#step-12-configure-environment)
13. [Setup Database](#step-13-setup-database)
14. [Run Migrations](#step-14-run-migrations)
15. [Build the Application](#step-15-build-the-application)
16. [Setup systemd Service](#step-16-setup-systemd-service)
17. [Configure Nginx](#step-17-configure-nginx)
18. [Install SSL Certificate](#step-18-install-ssl-certificate)
19. [Configure GitHub CI/CD](#step-19-configure-github-cicd)
20. [Verify Deployment](#step-20-verify-deployment)

---

## STEP 1: Create EC2 Instance

### In AWS Console (https://console.aws.amazon.com)

1. Go to **EC2 → Instances → Launch instances**

2. Configure:
   ```
   Name:           locolive-backend
   AMI:            Ubuntu Server 22.04 LTS (HVM), SSD Volume Type
   Architecture:   64-bit (x86)
   Instance type:  t3.micro
   Key pair:       locolive (your existing locolive.pem)
   ```

3. **Network settings** → Click "Edit"
   - Auto-assign public IP: **Enable**

4. **Configure storage**: 20 GB gp3

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

4. Verify (on your Mac):
   ```bash
   nslookup launchit.co.in
   # Should show your EC2 IP
   ```

---

## STEP 4: SSH into EC2

```bash
# On your Mac - Terminal

# Navigate to where your key is stored
cd ~/Downloads  # or wherever locolive.pem is

# Set correct permissions (only first time)
chmod 400 locolive.pem

# Connect to EC2
ssh -i locolive.pem ubuntu@YOUR_EC2_PUBLIC_IP

# Example:
# ssh -i locolive.pem ubuntu@13.235.123.456
```

✅ You should see: `ubuntu@ip-xxx-xxx-xxx-xxx:~$`

**All following commands are run ON EC2 (after SSH)**

---

## STEP 5: Update System Packages

```bash
# Update package list
sudo apt update

# Upgrade all packages
sudo apt upgrade -y

# Install essential tools
sudo apt install -y curl wget git unzip build-essential
```

---

## STEP 6: Install Go

```bash
# Download Go 1.22
wget https://go.dev/dl/go1.22.5.linux-amd64.tar.gz

# Remove any existing Go installation
sudo rm -rf /usr/local/go

# Extract to /usr/local
sudo tar -C /usr/local -xzf go1.22.5.linux-amd64.tar.gz

# Add to PATH (add to .bashrc for persistence)
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
# Expected: go version go1.22.5 linux/amd64

# Cleanup
rm go1.22.5.linux-amd64.tar.gz
```

---

## STEP 7: Install PostgreSQL

```bash
# Install PostgreSQL 16
sudo apt install -y postgresql postgresql-contrib

# Start and enable PostgreSQL
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Verify it's running
sudo systemctl status postgresql
# Should show: active (exited) - this is normal for PostgreSQL

# Verify version
psql --version
# Expected: psql (PostgreSQL) 14.x or higher
```

---

## STEP 8: Install Redis

```bash
# Install Redis
sudo apt install -y redis-server

# Configure Redis to start on boot
sudo sed -i 's/supervised no/supervised systemd/' /etc/redis/redis.conf

# Restart Redis to apply changes
sudo systemctl restart redis-server
sudo systemctl enable redis-server

# Verify it's running
sudo systemctl status redis-server
# Should show: active (running)

# Test Redis
redis-cli ping
# Expected: PONG
```

---

## STEP 9: Install Nginx

```bash
# Install Nginx
sudo apt install -y nginx

# Start and enable Nginx
sudo systemctl start nginx
sudo systemctl enable nginx

# Verify it's running
sudo systemctl status nginx
# Should show: active (running)

# Test in browser: http://YOUR_EC2_IP
# Should show "Welcome to nginx!" page
```

---

## STEP 10: Install Migrate Tool

```bash
# Download golang-migrate
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz

# Move to system bin
sudo mv migrate /usr/local/bin/migrate

# Make executable
sudo chmod +x /usr/local/bin/migrate

# Verify
migrate --version
# Expected: 4.17.0
```

---

## STEP 11: Clone Repository

### First, push your code to GitHub (on your Mac):

```bash
# On your Mac
cd /Users/pushp314/Desktop/LocoLiv/locolive-backend

# Initialize git if not already
git init

# Add all files
git add .

# Commit
git commit -m "LocoLive Backend - Native deployment"

# Create repo at https://github.com/new (name: locolive-backend)
# Then:
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

# Verify files
ls -la
# Should see: cmd/, internal/, deploy/, db/, etc.
```

---

## STEP 12: Configure Environment

```bash
# On EC2
cd /opt/locolive

# Copy production environment template
cp deploy/.env.production .env

# Generate a secure JWT secret
openssl rand -base64 64
# COPY THIS OUTPUT!

# Generate a secure database password
openssl rand -base64 32
# COPY THIS OUTPUT!

# Edit the environment file
nano .env
```

### Update `.env` with your values:

```env
# Server
PORT=8080
ENV=production

# Database - Use your generated password
DATABASE_URL=postgres://locolive:YOUR_DB_PASSWORD_HERE@localhost:5432/locolive?sslmode=disable
DB_HOST=localhost
DB_PORT=5432
DB_USER=locolive
DB_PASSWORD=YOUR_DB_PASSWORD_HERE
DB_NAME=locolive

# Redis
REDIS_URL=redis://localhost:6379

# JWT - Paste your generated secret
JWT_SECRET=YOUR_64_CHAR_SECRET_HERE
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# Google OAuth
GOOGLE_CLIENT_ID=your-google-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-google-client-secret

# Logging
LOG_LEVEL=info
```

**Save: `Ctrl+O`, `Enter`, `Ctrl+X`**

---

## STEP 13: Setup Database

```bash
# Switch to postgres user
sudo -u postgres psql

# Inside PostgreSQL shell, run these commands:
```

```sql
-- Create user (use same password as in .env)
CREATE USER locolive WITH PASSWORD 'YOUR_DB_PASSWORD_HERE';

-- Create database
CREATE DATABASE locolive;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE locolive TO locolive;

-- Connect to the database
\c locolive

-- Grant schema privileges
GRANT ALL ON SCHEMA public TO locolive;

-- Exit
\q
```

```bash
# Test the connection
psql -h localhost -U locolive -d locolive -c "SELECT 1;"
# Enter password when prompted
# Expected: Shows "1" in a table
```

---

## STEP 14: Run Migrations

```bash
# On EC2
cd /opt/locolive

# Run migrations
migrate -database "postgres://locolive:YOUR_DB_PASSWORD@localhost:5432/locolive?sslmode=disable" -path db/migrations up

# Expected output:
# 1/u init_schema (XXms)

# Verify tables were created
psql -h localhost -U locolive -d locolive -c "\dt"
# Should show: users, sessions, refresh_tokens, password_reset_tokens, schema_migrations
```

---

## STEP 15: Build the Application

```bash
# On EC2
cd /opt/locolive

# Download Go dependencies
go mod download

# Build the binary
go build -ldflags="-w -s" -o bin/api ./cmd/api

# Make it executable
chmod +x bin/api

# Test run (Ctrl+C to stop)
./bin/api
# Should show: "Server listening" on :8080

# Test in another terminal (or press Ctrl+C first)
curl http://localhost:8080/health
# Expected: {"status":"ok"...}
```

---

## STEP 16: Setup systemd Service

```bash
# On EC2

# Copy service file
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

# View logs
sudo journalctl -u locolive -f
# Press Ctrl+C to exit logs

# Test API
curl http://localhost:8080/health
```

---

## STEP 17: Configure Nginx

```bash
# On EC2

# Copy Nginx config
sudo cp /opt/locolive/deploy/nginx.conf /etc/nginx/sites-available/locolive

# Enable the site
sudo ln -sf /etc/nginx/sites-available/locolive /etc/nginx/sites-enabled/

# Remove default site
sudo rm -f /etc/nginx/sites-enabled/default

# Test configuration
sudo nginx -t
# Expected: "syntax is ok" and "test is successful"

# Reload Nginx
sudo systemctl reload nginx

# Test through Nginx
curl http://localhost/health
# Expected: {"status":"ok"...}

# Test from domain (if DNS is set)
curl http://launchit.co.in/health
```

---

## STEP 18: Install SSL Certificate

```bash
# On EC2

# Install Certbot
sudo apt install -y certbot python3-certbot-nginx

# Get SSL certificate
sudo certbot --nginx -d launchit.co.in -d www.launchit.co.in

# Follow prompts:
# 1. Enter your email address
# 2. Agree to terms (A)
# 3. Share email? (N)
# 4. Redirect HTTP to HTTPS? (2 - Redirect)

# Verify auto-renewal
sudo certbot renew --dry-run
# Should show: "Congratulations, all simulated renewals succeeded"

# Test HTTPS
curl https://launchit.co.in/health
# Expected: {"status":"ok"...}
```

---

## STEP 19: Configure GitHub CI/CD

### Add GitHub Secrets

1. Go to your GitHub repo → **Settings → Secrets and variables → Actions**

2. Add these secrets:

| Secret Name | Value |
|-------------|-------|
| `EC2_HOST` | Your EC2 public IP |
| `EC2_SSH_KEY` | Contents of locolive.pem |
| `DATABASE_URL` | `postgres://locolive:YOUR_PASSWORD@localhost:5432/locolive?sslmode=disable` |

### Get PEM content (on Mac):
```bash
cat ~/Downloads/locolive.pem
# Copy ENTIRE output including BEGIN/END lines
```

### Enable Actions:
1. Go to **Actions** tab
2. Enable workflows if prompted

Now every push to `main` → auto-deploy! 🎉

---

## STEP 20: Verify Deployment

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

## 📋 USEFUL COMMANDS

```bash
# === SSH ===
ssh -i ~/Downloads/locolive.pem ubuntu@YOUR_EC2_IP

# === Service Management ===
sudo systemctl status locolive    # Check status
sudo systemctl restart locolive   # Restart
sudo systemctl stop locolive      # Stop
sudo systemctl start locolive     # Start

# === View Logs ===
sudo journalctl -u locolive -f              # Live logs
sudo journalctl -u locolive --since "1 hour ago"  # Last hour
sudo tail -f /var/log/nginx/error.log       # Nginx errors

# === Database ===
psql -h localhost -U locolive -d locolive   # Connect to DB
sudo -u postgres psql                        # Admin access

# === Update & Redeploy (manual) ===
cd /opt/locolive
git pull origin main
go build -ldflags="-w -s" -o bin/api ./cmd/api
sudo systemctl restart locolive

# === Check Resources ===
free -h          # Memory
df -h            # Disk
htop             # CPU/Memory (install: sudo apt install htop)
```

---

## 🔧 TROUBLESHOOTING

### Service won't start
```bash
sudo journalctl -u locolive -n 50 --no-pager
```

### Database connection error
```bash
# Test connection
psql -h localhost -U locolive -d locolive -c "SELECT 1;"

# Check PostgreSQL is running
sudo systemctl status postgresql
```

### Nginx 502 Bad Gateway
```bash
# Check if API is running
curl http://localhost:8080/health

# Check systemd service
sudo systemctl status locolive
```

### Port already in use
```bash
sudo lsof -i :8080
sudo kill -9 <PID>
```

### Permission denied
```bash
sudo chown -R ubuntu:ubuntu /opt/locolive
chmod +x /opt/locolive/bin/api
```

---

## 💰 COST ESTIMATE

| Service | Monthly |
|---------|---------|
| EC2 t3.micro | ~$8.50 |
| EBS 20GB | ~$1.60 |
| Data Transfer | ~$2-5 |
| **Total** | **~$12-15** |

*First 12 months: 750 hrs/month free tier*

---

## ✅ DEPLOYMENT CHECKLIST

- [ ] EC2 instance created
- [ ] Security group configured (22, 80, 443)
- [ ] Domain DNS pointing to EC2
- [ ] SSH connected successfully
- [ ] System packages updated
- [ ] Go 1.22 installed
- [ ] PostgreSQL installed & running
- [ ] Redis installed & running
- [ ] Nginx installed & running
- [ ] Migrate tool installed
- [ ] Repository cloned
- [ ] `.env` configured
- [ ] Database user & database created
- [ ] Migrations executed
- [ ] Binary built
- [ ] systemd service running
- [ ] Nginx configured
- [ ] SSL certificate installed
- [ ] GitHub secrets configured
- [ ] All endpoints working

---

**🎉 Congratulations! Your LocoLive backend is live at https://launchit.co.in**
