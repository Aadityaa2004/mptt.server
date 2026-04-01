# Step 4: Deploy Services on Raspberry Pi

This guide walks you through deploying all services on your Raspberry Pi, including environment setup, pulling Docker images, and starting all containers.

## Prerequisites Checklist

Before starting, ensure you have completed:

- [x] **Step 1**: Built and pushed ARM64 Docker images to Docker Hub
- [x] **Step 2**: Configured DNS for `orpheus-networks.com`
- [x] **Step 3**: Set up router port forwarding (ports 80, 443, 8883, 9443)
- [ ] **Step 4**: Deploy services on Raspberry Pi ← **You are here**

### Additional Prerequisites

- Raspberry Pi 4/5 with Raspberry Pi OS (64-bit) installed
- Docker and Docker Compose installed
- Internet connection on Raspberry Pi
- SSH access to Raspberry Pi (or direct access via keyboard/monitor)

## Step-by-Step Deployment

### Step 1: Prepare Raspberry Pi

#### 1.1 Update System

```bash
# Update package lists
sudo apt update

# Upgrade system packages
sudo apt upgrade -y

# Reboot if kernel was updated
sudo reboot
```

#### 1.2 Install Docker (if not already installed)

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add your user to docker group (replace 'pi' with your username)
sudo usermod -aG docker $USER

# Install Docker Compose plugin
sudo apt install docker-compose-plugin -y

# Log out and back in for group changes to take effect
# Or run: newgrp docker

# Verify installation
docker --version
docker compose version
```

#### 1.3 Configure Firewall (Optional but Recommended)

```bash
# Install UFW if not installed
sudo apt install ufw -y

# Allow SSH (important - do this first!)
sudo ufw allow 22/tcp

# Allow HTTP/HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Allow MQTT ports
sudo ufw allow 8883/tcp
sudo ufw allow 9443/tcp

# Enable firewall
sudo ufw enable

# Check status
sudo ufw status
```

#### 1.4 Set Static IP (Recommended)

To prevent IP changes from breaking port forwarding:

```bash
# Edit network configuration
sudo nano /etc/dhcpcd.conf

# Add at the end (adjust for your network):
interface eth0  # Use 'wlan0' for WiFi
static ip_address=192.168.1.100/24  # Your desired IP
static routers=192.168.1.1  # Your router IP
static domain_name_servers=192.168.1.1 8.8.8.8 8.8.4.4

# Save and exit (Ctrl+X, Y, Enter)

# Restart networking
sudo systemctl restart dhcpcd

# Verify IP
hostname -I
```

### Step 2: Clone Repository and Setup

#### 2.1 Clone Repository

```bash
# Navigate to home directory
cd ~

# Clone your repository (replace with your repo URL)
git clone <your-repo-url> mptt-server
# Or if you have the files already, skip this step

# Navigate to project directory
cd mptt-server
```

**Alternative:** If you don't have a git repository, you can:
- Copy files via SCP from your development machine
- Use USB drive to transfer files
- Create files manually on Raspberry Pi

#### 2.2 Verify Files Are Present

```bash
# Check that deployment files exist
ls -la deploy-rpi.sh setup-env.sh setup-ssl.sh
ls -la docker-compose.rpi.yml docker-compose.nginx.yml
ls -la env.production.example
```

### Step 3: Configure Environment Variables

#### 3.1 Create Environment File

You have two options:

**Option A: Interactive Setup (Recommended)**

```bash
# Run the interactive setup script
./setup-env.sh
```

This script will:
- Copy the template file
- Prompt you for all required values
- Generate secure random secrets
- Set proper file permissions

**Option B: Manual Setup**

```bash
# Copy template file
cp env.production.example .env.production

# Edit the file
nano .env.production
```

#### 3.2 Required Values to Configure

Edit `.env.production` and replace all `CHANGE_ME` values:

**Domain Configuration:**
```bash
DOMAIN=orpheus-networks.com
# Origin only — do not append /api (the app adds /api/... to every API path).
NEXT_PUBLIC_API_BASE_URL=https://orpheus-networks.com
NEXT_PUBLIC_READINGS_API_BASE_URL=https://orpheus-networks.com
CORS_ALLOWED_ORIGINS=https://orpheus-networks.com,https://www.orpheus-networks.com
```

**Docker Hub:**
```bash
DOCKERHUB_USERNAME=your-dockerhub-username
```

**Database:**
```bash
POSTGRES_USER=iot_user
POSTGRES_PASSWORD=your-strong-password-here
POSTGRES_DB=iot
```

**JWT Secrets (generate strong random values):**
```bash
# Generate with: openssl rand -base64 64
JWT_SECRET_KEY=your-generated-secret-here
INTERNAL_API_SECRET=your-generated-secret-here
```

**Admin Account:**
```bash
ADMIN_USERNAME=admin
ADMIN_EMAIL=admin@orpheus-networks.com
ADMIN_PASSWORD=your-strong-admin-password
```

**MQTT Broker:**
```bash
BROKER_USER=your-mqtt-username
BROKER_PASS=your-mqtt-password
```

**External APIs (if used):**
```bash
OPENWEATHER_API_KEY=your-api-key-if-needed
```

#### 3.3 Set File Permissions

```bash
# Set secure permissions (owner read/write only)
chmod 600 .env.production

# Verify permissions
ls -la .env.production
# Should show: -rw------- (600)
```

#### 3.4 Generate Secure Secrets

If you need to generate secrets:

```bash
# Generate JWT secret (64 characters)
openssl rand -base64 64

# Generate password (32 characters)
openssl rand -base64 32

# Generate shorter secret (24 characters)
openssl rand -base64 24
```

### Step 4: Login to Docker Hub

```bash
# Login to Docker Hub (needed to pull images)
docker login

# Enter your Docker Hub username and password
```

**Note:** If you have 2FA enabled, you'll need to use an access token instead of password.

### Step 5: Deploy Services

#### 5.1 Run Deployment Script

```bash
# Make sure script is executable
chmod +x deploy-rpi.sh

# Run deployment
./deploy-rpi.sh
```

The script will:
1. Validate `.env.production` file exists and has correct permissions
2. Check required environment variables are set
3. Pull Docker images from Docker Hub
4. Create Docker network
5. Start all services
6. Verify deployment

**Expected output:**
```
========================================
Raspberry Pi Production Deployment
========================================

✓ Environment file validated

Pulling Docker images from Docker Hub...

✓ Images pulled successfully

Setting up Docker network...
✓ Network ready

Starting services...

✓ Services started

Checking service health...
[Container status table]

========================================
Deployment completed successfully!
========================================
```

#### 5.2 Manual Deployment (Alternative)

If the script doesn't work, deploy manually:

```bash
# Load environment variables
set -a
source .env.production
set +a

# Create network
docker network create mqtt-network 2>/dev/null || true

# Pull images
docker compose -f docker-compose.rpi.yml pull

# Start services
docker compose -f docker-compose.rpi.yml up -d

# Check status
docker compose -f docker-compose.rpi.yml ps
```

### Step 6: Verify Services Are Running

#### 6.1 Check Container Status

```bash
# List all containers
docker ps

# Check specific services
docker ps | grep -E '(ingestor|api|frontend|postgres|mosquitto|bridge)'

# View logs
docker compose -f docker-compose.rpi.yml logs -f
```

**Expected:** All containers should show status "Up" or "Up (healthy)".

#### 6.2 Test Services Locally

```bash
# Test API health endpoint
curl http://localhost:9002/health

# Test frontend (should return HTML)
curl http://localhost:3000

# Test MQTT broker
docker exec mqtt-broker-prod mosquitto_pub -h localhost -t test -m "test"
```

#### 6.3 Check Service Logs

```bash
# View all logs
docker compose -f docker-compose.rpi.yml logs

# View specific service logs
docker logs api-service-prod
docker logs mqt-frontend-prod
docker logs mqtt-broker-prod
docker logs postgresql-prod

# Follow logs in real-time
docker compose -f docker-compose.rpi.yml logs -f api-service
```

### Step 7: Start Nginx Reverse Proxy

**Important:** Don't start Nginx until SSL certificates are set up (Step 5). However, you can prepare it:

```bash
# Create network if it doesn't exist (should already exist from Step 5)
docker network create mqtt-network 2>/dev/null || true

# Start Nginx (will fail SSL until certificates are set up)
docker compose -f docker-compose.nginx.yml up -d

# Check Nginx status
docker logs nginx-proxy
```

**Note:** Nginx will start but won't serve HTTPS until SSL certificates are configured in Step 5.

### Step 8: Verify External Access

#### 8.1 Test from External Network

1. **Disconnect from your WiFi** (use mobile data or different network)
2. **Test HTTP access:**
   ```bash
   curl http://orpheus-networks.com
   # or visit in browser: http://orpheus-networks.com
   ```

3. **Test API (liveness):**
   ```bash
   curl http://orpheus-networks.com/health/live
   ```

**Expected:** Should connect to your Raspberry Pi (may see errors until SSL is set up).

#### 8.2 Test from Local Network

```bash
# Test using local IP
curl http://192.168.1.100:9002/health  # Replace with your Pi's IP
curl http://192.168.1.100:3000
```

## Troubleshooting

### Services Won't Start

**Problem:** Containers fail to start

**Solutions:**
1. **Check logs:**
   ```bash
   docker compose -f docker-compose.rpi.yml logs
   ```

2. **Verify environment file:**
   ```bash
   # Check file exists and has values
   cat .env.production | grep -v "^#" | grep -v "^$"
   
   # Check permissions
   ls -la .env.production
   ```

3. **Check Docker Hub login:**
   ```bash
   docker login
   ```

4. **Verify images exist:**
   ```bash
   docker images | grep mptt-server
   ```

### "Image not found" or "Pull access denied"

**Problem:** Can't pull images from Docker Hub

**Solutions:**
1. **Login to Docker Hub:**
   ```bash
   docker login
   ```

2. **Verify Docker Hub username in .env.production:**
   ```bash
   grep DOCKERHUB_USERNAME .env.production
   ```

3. **Check images exist on Docker Hub:**
   - Visit: https://hub.docker.com
   - Search for your images: `your-username/mptt-server-ingestor`

4. **Verify images are ARM64:**
   ```bash
   docker buildx imagetools inspect your-username/mptt-server-ingestor:latest
   ```

### Database Connection Errors

**Problem:** Services can't connect to PostgreSQL

**Solutions:**
1. **Check PostgreSQL is running:**
   ```bash
   docker ps | grep postgres
   docker logs postgresql-prod
   ```

2. **Verify credentials in .env.production:**
   ```bash
   grep POSTGRES .env.production
   ```

3. **Test database connection:**
   ```bash
   docker exec postgresql-prod psql -U iot_user -d iot -c "SELECT 1;"
   ```

### Port Already in Use

**Problem:** Port conflict errors

**Solutions:**
1. **Check what's using the port:**
   ```bash
   sudo netstat -tuln | grep -E ':(80|443|9002|3000)'
   # or
   sudo ss -tuln | grep -E ':(80|443|9002|3000)'
   ```

2. **Stop conflicting services:**
   ```bash
   # If Apache is running
   sudo systemctl stop apache2
   
   # If Nginx is running (system nginx)
   sudo systemctl stop nginx
   ```

3. **Change ports in docker-compose if needed** (not recommended)

### Out of Disk Space

**Problem:** Docker runs out of space

**Solutions:**
1. **Check disk space:**
   ```bash
   df -h
   ```

2. **Clean up Docker:**
   ```bash
   # Remove unused images
   docker image prune -a
   
   # Remove unused volumes
   docker volume prune
   
   # Remove everything unused
   docker system prune -a
   ```

3. **Expand filesystem** (if using Raspberry Pi OS):
   ```bash
   sudo raspi-config
   # Advanced Options → Expand Filesystem
   ```

### Services Keep Restarting

**Problem:** Containers restart in a loop

**Solutions:**
1. **Check logs for errors:**
   ```bash
   docker logs <container-name> --tail 100
   ```

2. **Check health checks:**
   ```bash
   docker inspect <container-name> | grep -A 10 Health
   ```

3. **Temporarily disable health checks** to see actual errors:
   - Comment out healthcheck in docker-compose.rpi.yml
   - Restart services

### Network Issues

**Problem:** Containers can't communicate

**Solutions:**
1. **Check Docker network:**
   ```bash
   docker network ls
   docker network inspect mqtt-network
   ```

2. **Recreate network:**
   ```bash
   docker network rm mqtt-network
   docker network create mqtt-network
   docker compose -f docker-compose.rpi.yml up -d
   ```

## Post-Deployment Checklist

- [ ] All containers are running (`docker ps` shows all services)
- [ ] Environment file configured correctly
- [ ] Services respond locally (test with curl)
- [ ] Docker Hub images pulled successfully
- [ ] No errors in logs
- [ ] Database is accessible
- [ ] API health endpoint responds
- [ ] Frontend is accessible
- [ ] MQTT broker is running
- [ ] External access works (from different network)

## Next Steps

After services are deployed and running:

1. ✅ **Step 4 Complete** - Services are running on Raspberry Pi
2. **Step 5** - Set up SSL certificates with Let's Encrypt
3. **Step 6** - Start Nginx reverse proxy
4. **Step 7** - Final verification and testing

## Useful Commands Reference

```bash
# View all running containers
docker ps

# View all containers (including stopped)
docker ps -a

# View logs
docker compose -f docker-compose.rpi.yml logs -f

# Restart services
docker compose -f docker-compose.rpi.yml restart

# Stop services
docker compose -f docker-compose.rpi.yml down

# Start services
docker compose -f docker-compose.rpi.yml up -d

# Update and restart
docker compose -f docker-compose.rpi.yml pull
docker compose -f docker-compose.rpi.yml up -d

# View resource usage
docker stats

# Execute command in container
docker exec -it api-service-prod sh
docker exec -it postgresql-prod psql -U iot_user -d iot
```

## Important Notes

- **Keep .env.production secure** - Never commit it to git
- **Backup your configuration** - Save .env.production in a secure location
- **Monitor logs** - Check logs regularly for errors
- **Keep system updated** - Run `sudo apt update && sudo apt upgrade` regularly
- **Monitor disk space** - Docker can use significant disk space

Once all services are running successfully, proceed to Step 5: SSL Certificate Setup!

