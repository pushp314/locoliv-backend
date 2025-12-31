# 🚀 Easy AWS Deployment Guide
## LocoLive Backend → launchit.co.in

I have created an **automated installer** (`deploy/install.sh`) that does almost everything for you.

---

## Step 1: Push Code to GitHub (On your Mac)

```bash
cd /Users/pushp314/Desktop/LocoLiv/locolive-backend

# Make script executable
chmod +x deploy/install.sh

git add .
git commit -m "Add installer script"
git push origin main
```

---

## Step 2: Create Server & Connect

1. **Launch EC2 Instance** (Ubuntu 22.04 t3.micro)
   - Allow **HTTP, HTTPS, SSH** traffic.
   - Use key: `locolive.pem`.

2. **Connect via SSH** (Terminal on Mac):
```bash
cd ~/Downloads
chmod 400 locolive.pem
ssh -i locolive.pem ubuntu@YOUR_EC2_IP
```

---

## Step 3: Run the Auto-Installer (On EC2)

Run these 4 commands on your server:

```bash
# 1. Clone your repo
git clone https://github.com/YOUR_USERNAME/locolive-backend.git

# 2. Go to deploy folder
cd locolive-backend/deploy

# 3. Make script executable
chmod +x install.sh

# 4. Run it!
sudo ./install.sh
```

**That's it!** The script will:
- ✅ Install Go, Postgres, Redis, Nginx
- ✅ Create database & secure passwords automatically
- ✅ Configure systemd & start server
- ✅ Ask you for Google keys (optional, can skip)

---

## Step 4: Finish SSL (HTTPS)

The script stops just before SSL to let you enter your email interactively. Run:

```bash
sudo certbot --nginx -d launchit.co.in
```

---

## 🔧 Useful Commands

```bash
# View Logs
sudo journalctl -u locolive -f

# Restart Server
sudo systemctl restart locolive

# Edit Config
sudo nano /opt/locolive/.env
```
