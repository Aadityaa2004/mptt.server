# Quick Start Guide - Raspberry Pi Deployment

This is a quick reference for deploying on Raspberry Pi. For detailed instructions, see [DEPLOYMENT.md](DEPLOYMENT.md).

## Prerequisites Checklist

- [ ] Raspberry Pi 4/5 with Raspberry Pi OS (64-bit)
- [ ] Docker and Docker Compose installed
- [ ] Domain `orpheus-networks.com` DNS pointing to your public IP
- [ ] Router ports forwarded: 80, 443, 8883, 9443
- [ ] Docker Hub images built and pushed (see [BUILD_ARM64.md](BUILD_ARM64.md))

## Quick Deployment Steps

### 1. Clone and Setup

```bash
git clone <your-repo>
cd mptt.server
```

### 2. Configure Environment

```bash
# Option A: Interactive setup
./setup-env.sh

# Option B: Manual setup
cp env.production.example .env.production
nano .env.production  # Fill in all CHANGE_ME values
chmod 600 .env.production
```

**Required values to change:**
- `DOCKERHUB_USERNAME` - Your Docker Hub username
- `POSTGRES_PASSWORD` - Strong password
- `JWT_SECRET_KEY` - Generate with: `openssl rand -base64 64`
- `INTERNAL_API_SECRET` - Generate with: `openssl rand -base64 32`
- `ADMIN_PASSWORD` - Strong admin password
- `BROKER_USER` and `BROKER_PASS` - MQTT credentials

### 3. Deploy Services

```bash
./deploy-rpi.sh
```

### 4. Setup SSL

```bash
./setup-ssl.sh
```

### 5. Start Nginx

```bash
docker compose -f docker-compose.nginx.yml up -d
```

### 6. Verify

```bash
# Check services
docker ps

# Test API
curl https://orpheus-networks.com/api/health

# Access in browser
# https://orpheus-networks.com
```

## Common Commands

```bash
# View logs
docker compose -f docker-compose.rpi.yml logs -f

# Stop services
docker compose -f docker-compose.rpi.yml down
docker compose -f docker-compose.nginx.yml down

# Start services
docker compose -f docker-compose.rpi.yml up -d
docker compose -f docker-compose.nginx.yml up -d

# Update images
docker compose -f docker-compose.rpi.yml pull
docker compose -f docker-compose.rpi.yml up -d
```

## Troubleshooting

**Services won't start:**
- Check `.env.production` exists and has correct values
- Check file permissions: `chmod 600 .env.production`
- Check logs: `docker compose -f docker-compose.rpi.yml logs`

**SSL certificate fails:**
- Verify DNS: `dig orpheus-networks.com`
- Check port 80 is accessible from internet
- Check firewall allows port 80

**Can't access website:**
- Verify DNS propagation
- Check router port forwarding
- Check Nginx logs: `docker logs nginx-proxy`

For more details, see [DEPLOYMENT.md](DEPLOYMENT.md).

