# Raspberry Pi Production Deployment Guide

This guide walks you through deploying the MQTT server application on a Raspberry Pi with production-ready configuration, SSL/TLS, and global internet access via the domain `orpheus-networks.com`.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Network Configuration](#network-configuration)
3. [Domain Setup](#domain-setup)
4. [Environment Configuration](#environment-configuration)
5. [SSL Certificate Setup](#ssl-certificate-setup)
6. [Deployment](#deployment)
7. [Verification](#verification)
8. [Maintenance](#maintenance)
9. [Troubleshooting](#troubleshooting)

## Prerequisites

### Hardware Requirements

- Raspberry Pi 4 or 5 (ARM64 architecture)
- At least 4GB RAM recommended
- MicroSD card (32GB+ recommended)
- Stable internet connection
- Power supply (official Raspberry Pi power adapter recommended)

### Software Requirements

- Raspberry Pi OS (64-bit) installed and updated
- Docker installed (version 20.10+)
- Docker Compose installed (version 2.0+)
- Domain name: `orpheus-networks.com` (or your domain)
- Docker Hub account with images pushed

### Install Docker and Docker Compose

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo apt install docker-compose-plugin -y

# Log out and back in for group changes to take effect
# Or run: newgrp docker

# Verify installation
docker --version
docker compose version
```

## Network Configuration

### Router Port Forwarding

Configure your router to forward the following ports to your Raspberry Pi's local IP address:

| Port | Protocol | Service | Description |
|------|----------|---------|-------------|
| 80 | TCP | HTTP | Nginx (redirects to HTTPS) |
| 443 | TCP | HTTPS | Nginx (SSL/TLS) |
| 8883 | TCP | MQTT TLS | MQTT Broker (for IoT devices) |
| 9443 | TCP | MQTT WS TLS | MQTT WebSocket (for web clients) |

**Steps:**

1. Log into your router's admin panel (usually `192.168.1.1` or `192.168.0.1`)
2. Navigate to "Port Forwarding" or "Virtual Server" settings
3. Add rules for each port above, forwarding to your Raspberry Pi's local IP
4. Save and apply changes

### Firewall Configuration

If you're using `ufw` (Uncomplicated Firewall):

```bash
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp     # HTTP
sudo ufw allow 443/tcp    # HTTPS
sudo ufw allow 8883/tcp   # MQTT TLS
sudo ufw allow 9443/tcp   # MQTT WebSocket TLS
sudo ufw enable
```

### Find Your Raspberry Pi's IP Address

```bash
hostname -I
# or
ip addr show
```

## Domain Setup

### DNS Configuration

1. Log into your domain registrar (where you purchased `orpheus-networks.com`)
2. Navigate to DNS management
3. Add/Update the following DNS records:

**A Record (IPv4):**
- **Name:** `@` or `orpheus-networks.com`
- **Type:** A
- **Value:** Your public IP address
- **TTL:** 3600 (or default)

**A Record (IPv4) - WWW:**
- **Name:** `www`
- **Type:** A
- **Value:** Your public IP address
- **TTL:** 3600 (or default)

**Find Your Public IP:**
```bash
curl ifconfig.me
# or visit: https://whatismyipaddress.com
```

4. Wait for DNS propagation (can take up to 48 hours, usually much faster)

**Verify DNS:**
```bash
dig orpheus-networks.com
# or
nslookup orpheus-networks.com
```

## Environment Configuration

### Step 1: Create Environment File

```bash
# Copy the example file
cp .env.production.example .env.production

# Or use the interactive setup script
./setup-env.sh
```

### Step 2: Edit Environment Variables

Open `.env.production` and fill in all values marked with `CHANGE_ME`:

```bash
nano .env.production
```

**Required Variables:**

- `DOCKERHUB_USERNAME` - Your Docker Hub username
- `POSTGRES_PASSWORD` - Strong database password
- `JWT_SECRET_KEY` - Strong random secret (64+ characters)
- `INTERNAL_API_SECRET` - Strong random secret for service-to-service auth
- `ADMIN_PASSWORD` - Strong admin account password
- `BROKER_USER` - MQTT broker username
- `BROKER_PASS` - MQTT broker password

**Generate Strong Secrets:**

```bash
# Generate JWT secret
openssl rand -base64 64

# Generate password
openssl rand -base64 32
```

### Step 3: Set File Permissions

```bash
chmod 600 .env.production
```

**Important:** Never commit `.env.production` to git! It contains sensitive secrets.

## SSL Certificate Setup

### Automatic Setup (Recommended)

```bash
./setup-ssl.sh
```

This script will:
1. Start Nginx container
2. Request SSL certificate from Let's Encrypt
3. Configure automatic renewal
4. Set up cron job for certificate renewal

### Manual Setup

If automatic setup fails:

```bash
# Start Nginx first
docker compose -f docker-compose.nginx.yml up -d

# Request certificate
docker exec nginx-proxy certbot certonly \
    --webroot \
    --webroot-path=/var/www/certbot \
    --email admin@orpheus-networks.com \
    --agree-tos \
    --no-eff-email \
    -d orpheus-networks.com \
    -d www.orpheus-networks.com

# Reload Nginx
docker exec nginx-proxy nginx -s reload
```

## Deployment

### Step 1: Clone Repository (if not already done)

```bash
git clone <your-repo-url>
cd mptt.server
```

### Step 2: Configure Environment

```bash
# Use interactive setup or manually edit
./setup-env.sh
```

### Step 3: Deploy Services

```bash
./deploy-rpi.sh
```

This script will:
1. Validate environment file
2. Pull Docker images from Docker Hub
3. Create Docker network
4. Start all services
5. Verify health

### Step 4: Start Nginx Reverse Proxy

```bash
docker compose -f docker-compose.nginx.yml up -d
```

### Step 5: Verify Deployment

```bash
# Check running containers
docker ps

# Check logs
docker compose -f docker-compose.rpi.yml logs -f

# Check specific service logs
docker logs api-service-prod
docker logs mqt-frontend-prod
docker logs nginx-proxy
```

## Verification

### Health Checks

```bash
# API Health
curl https://orpheus-networks.com/api/health

# Frontend
curl -I https://orpheus-networks.com
```

### Browser Testing

1. Open `https://orpheus-networks.com` in your browser
2. Verify SSL certificate is valid (green padlock)
3. Test login functionality
4. Verify API endpoints are accessible

### SSL Certificate Check

Visit: https://www.ssllabs.com/ssltest/analyze.html?d=orpheus-networks.com

## Maintenance

### View Logs

```bash
# All services
docker compose -f docker-compose.rpi.yml logs -f

# Specific service
docker logs api-service-prod -f
docker logs mqt-frontend-prod -f
docker logs nginx-proxy -f
```

### Update Services

```bash
# Pull latest images
docker compose -f docker-compose.rpi.yml pull

# Restart services
docker compose -f docker-compose.rpi.yml up -d
```

### Backup Database

```bash
# Create backup
docker exec postgresql-prod pg_dump -U iot_user iot > backup_$(date +%Y%m%d_%H%M%S).sql

# Restore backup
cat backup_file.sql | docker exec -i postgresql-prod psql -U iot_user iot
```

### Certificate Renewal

Certificates automatically renew via cron job. To manually renew:

```bash
docker exec nginx-proxy certbot renew
docker exec nginx-proxy nginx -s reload
```

### Stop Services

```bash
# Stop application services
docker compose -f docker-compose.rpi.yml down

# Stop Nginx
docker compose -f docker-compose.nginx.yml down
```

### Start Services

```bash
# Start application services
docker compose -f docker-compose.rpi.yml up -d

# Start Nginx
docker compose -f docker-compose.nginx.yml up -d
```

## Troubleshooting

### Services Won't Start

1. **Check Docker:**
   ```bash
   docker ps
   docker info
   ```

2. **Check Logs:**
   ```bash
   docker compose -f docker-compose.rpi.yml logs
   ```

3. **Check Environment:**
   ```bash
   # Verify .env.production exists and has correct permissions
   ls -la .env.production
   ```

### SSL Certificate Issues

1. **Certificate Not Generated:**
   - Verify DNS points to your server: `dig orpheus-networks.com`
   - Check port 80 is accessible: `curl http://orpheus-networks.com`
   - Verify firewall allows port 80

2. **Certificate Expired:**
   ```bash
   # Check certificate expiry
   docker exec nginx-proxy certbot certificates
   
   # Manually renew
   docker exec nginx-proxy certbot renew --force-renewal
   docker exec nginx-proxy nginx -s reload
   ```

### Database Connection Issues

1. **Check PostgreSQL:**
   ```bash
   docker logs postgresql-prod
   docker exec postgresql-prod pg_isready -U iot_user
   ```

2. **Verify Credentials:**
   - Check `.env.production` has correct `POSTGRES_USER` and `POSTGRES_PASSWORD`

### Network Issues

1. **Check Docker Network:**
   ```bash
   docker network ls
   docker network inspect mqtt-network
   ```

2. **Verify Port Forwarding:**
   - Test from external network: `curl http://your-public-ip`
   - Check router port forwarding rules

3. **Check Firewall:**
   ```bash
   sudo ufw status
   sudo iptables -L -n
   ```

### Frontend Not Loading

1. **Check Frontend Container:**
   ```bash
   docker logs mqt-frontend-prod
   docker exec mqt-frontend-prod curl http://localhost:3000
   ```

2. **Check Nginx Configuration:**
   ```bash
   docker exec nginx-proxy nginx -t
   docker logs nginx-proxy
   ```

3. **Verify API URL:**
   - Check `.env.production` has correct `NEXT_PUBLIC_API_BASE_URL`

### API Not Responding

1. **Check API Container:**
   ```bash
   docker logs api-service-prod
   docker exec api-service-prod curl http://localhost:9002/health
   ```

2. **Check Database Connection:**
   ```bash
   docker exec api-service-prod env | grep POSTGRES
   ```

### MQTT Broker Issues

1. **Check Mosquitto:**
   ```bash
   docker logs mqtt-broker-prod
   docker exec mqtt-broker-prod mosquitto_pub -h localhost -t test -m "test"
   ```

2. **Verify Ports:**
   ```bash
   netstat -tuln | grep 8883
   ```

## Security Best Practices

1. **Keep System Updated:**
   ```bash
   sudo apt update && sudo apt upgrade -y
   ```

2. **Use Strong Passwords:**
   - Generate random passwords for all services
   - Use password manager

3. **Restrict File Permissions:**
   ```bash
   chmod 600 .env.production
   ```

4. **Regular Backups:**
   - Set up automated database backups
   - Backup configuration files

5. **Monitor Logs:**
   - Regularly check service logs
   - Set up log rotation

6. **Firewall Rules:**
   - Only expose necessary ports
   - Use fail2ban for SSH protection

## Support

For issues or questions:
1. Check logs: `docker compose -f docker-compose.rpi.yml logs`
2. Review this documentation
3. Check GitHub issues (if applicable)

## Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Nginx Documentation](https://nginx.org/en/docs/)
- [Let's Encrypt Documentation](https://letsencrypt.org/docs/)
- [Raspberry Pi Documentation](https://www.raspberrypi.org/documentation/)

