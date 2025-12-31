# AWS EC2 Deployment Guide - LocoLive Backend

Complete step-by-step guide to deploy the LocoLive Go backend to AWS EC2 Ubuntu t3.micro.

---

## Prerequisites

1. **AWS Account** with EC2 access
2. **SSH Key Pair** (`.pem` file) for EC2 access
3. **Google OAuth credentials** (Client ID & Secret)

---

## Step 1: Launch EC2 Instance

### In AWS Console:

1. Go to **EC2 → Launch Instance**
2. Configure:
   - **Name**: `locolive-backend`
   - **AMI**: Ubuntu 22.04 LTS
   - **Instance type**: t3.micro (free tier eligible)
   - **Key pair**: Select or create one
   - **Security Group**: Create with these rules:

| Type | Port | Source |
|------|------|--------|
| SSH | 22 | Your IP |
| HTTP | 80 | 0.0.0.0/0 |
| HTTPS | 443 | 0.0.0.0/0 |

3. **Launch** the instance
4. Note the **Public IP** address

---

## Step 2: Connect to EC2

```bash
# From your Mac
chmod 400 your-key.pem
ssh -i your-key.pem ubuntu@YOUR_EC2_PUBLIC_IP
```

---

## Step 3: Setup EC2 (Run on EC2)

```bash
# Download and run setup script
curl -O https://raw.githubusercontent.com/YOUR_REPO/deploy/setup-ec2.sh
chmod +x setup-ec2.sh
./setup-ec2.sh

# Log out and back in for Docker permissions
exit
```

Then SSH back in.

---

## Step 4: Copy Project Files

### From your Mac (new terminal):

```bash
cd /Users/pushp314/Desktop/LocoLiv/locolive-backend

# Create archive (excluding unnecessary files)
tar --exclude='bin' --exclude='.git' --exclude='node_modules' \
    -czvf locolive-backend.tar.gz .

# Copy to EC2
scp -i your-key.pem locolive-backend.tar.gz ubuntu@YOUR_EC2_IP:/tmp/

# SSH into EC2
ssh -i your-key.pem ubuntu@YOUR_EC2_IP
```

### On EC2:

```bash
# Extract files
cd /opt/locolive
tar -xzvf /tmp/locolive-backend.tar.gz
rm /tmp/locolive-backend.tar.gz
```

---

## Step 5: Configure Environment

```bash
# On EC2
cd /opt/locolive

# Copy production env template
cp deploy/.env.production .env

# Edit with your values
nano .env
```

### Required changes in `.env`:

```env
# Generate a secure JWT secret (run this on your Mac)
# openssl rand -base64 64

JWT_SECRET=paste_your_64_char_secret_here
DB_PASSWORD=your_secure_db_password
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
```

**Important**: Update `DATABASE_URL` with your `DB_PASSWORD`:
```env
DATABASE_URL=postgres://locolive:YOUR_DB_PASSWORD@postgres:5432/locolive?sslmode=disable
```

---

## Step 6: Deploy!

```bash
cd /opt/locolive
chmod +x deploy/deploy.sh
./deploy/deploy.sh
```

This will:
- Build and start Docker containers
- Run database migrations
- Configure Nginx
- Enable auto-start on boot

---

## Step 7: Verify Deployment

```bash
# Test locally on EC2
curl http://localhost:8080/health

# Test from outside (your Mac)
curl http://YOUR_EC2_PUBLIC_IP/health
```

Expected response:
```json
{"status":"ok","timestamp":"...","version":"1.0.0"}
```

---

## Useful Commands

```bash
# View logs
docker-compose -f deploy/docker-compose.prod.yml logs -f

# Restart API
sudo systemctl restart locolive

# Check status
docker-compose -f deploy/docker-compose.prod.yml ps

# Run migrations again
docker-compose -f deploy/docker-compose.prod.yml --profile migrate run --rm migrate

# Stop everything
docker-compose -f deploy/docker-compose.prod.yml down
```

---

## Optional: Setup SSL (HTTPS)

```bash
# Install Certbot
sudo apt install -y certbot python3-certbot-nginx

# Get SSL certificate (replace with your domain)
sudo certbot --nginx -d your-domain.com

# Auto-renewal is configured automatically
```

---

## Cost Estimate (t3.micro)

| Service | Monthly Cost |
|---------|-------------|
| EC2 t3.micro | ~$8.50 |
| EBS 8GB | ~$0.80 |
| Data transfer | ~$1-5 |
| **Total** | **~$10-15/mo** |

*Free tier: 750 hours/month for first 12 months*

---

## Troubleshooting

### Container won't start
```bash
docker-compose -f deploy/docker-compose.prod.yml logs api
```

### Database connection failed
```bash
docker-compose -f deploy/docker-compose.prod.yml logs postgres
docker exec -it locolive-postgres-1 psql -U locolive -c "\l"
```

### Port 80 not accessible
- Check EC2 Security Group allows port 80
- Check Nginx status: `sudo systemctl status nginx`
