# Complete Raspberry Pi Deployment Steps (Cloudflare Tunnel)

This is your **step-by-step checklist** for deploying on a Raspberry Pi using Cloudflare Tunnel (campus / no port forwarding).

---

## Prerequisites

- [ ] Raspberry Pi 4/5 with Raspberry Pi OS (64-bit)
- [ ] Docker and Docker Compose installed
- [ ] Domain `orpheus-networks.com` added to Cloudflare
- [ ] Cloudflare Tunnel created and token obtained
- [ ] Docker Hub images pushed (see section below)

---

## Testing the RPi stack locally (optional)

To run the same stack on your **dev machine** with **localhost** (same nginx routing, no orpheus-networks.com):

1. `cp docker-compose.rpi.override.example.yml docker-compose.rpi.override.yml`
2. `cp env.rpi.local.example .env.rpi.local` and edit (or keep using `.env.production`; override will still set CORS and build frontend with localhost)
3. `docker compose -f docker-compose.rpi.yml -f docker-compose.rpi.override.yml --env-file .env.rpi.local up --build`
4. Open **http://localhost** (nginx). See [README.md](README.md) → "Testing the RPi stack locally" for details.

---

## Step 1: Install Docker & Docker Compose on Pi

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Install Docker Compose plugin
sudo apt install docker-compose-plugin -y

# Log out and back in (or run: newgrp docker)
# Verify installation
docker --version
docker compose version
```

---

## Step 2: Clone Repository on Pi

```bash
# Clone the repo (you need the compose files, nginx configs, scripts)
git clone <your-repo-url>
cd mptt.server
```

**Why clone?** You need:
- `docker-compose.rpi.yml` (defines all services)
- `nginx/` configs (routing rules)
- `deploy-rpi.sh` (deployment script)
- `env.production.example` (template)

**You do NOT rebuild images** - you'll pull them from Docker Hub.

---

## Step 3: Create `.env.production` File

```bash
# Copy the example file
cp env.production.example .env.production

# Edit it
nano .env.production
```

### Required Values to Set:

```bash
# Docker Hub Configuration
DOCKERHUB_USERNAME=maplesense
IMAGE_TAG=v1.0.0

# Cloudflare Tunnel Token (from Cloudflare dashboard)
CLOUDFLARE_TUNNEL_TOKEN=<your-token-here>

# Domain Configuration
DOMAIN=orpheus-networks.com
NEXT_PUBLIC_API_BASE_URL=https://orpheus-networks.com/api
NEXT_PUBLIC_READINGS_API_BASE_URL=https://orpheus-networks.com/api
CORS_ALLOWED_ORIGINS=https://orpheus-networks.com,https://www.orpheus-networks.com

# Database Configuration
POSTGRES_USER=iot_user
POSTGRES_PASSWORD=<strong-password>
POSTGRES_DB=iot
POSTGRES_SSLMODE=disable

# JWT Secrets (generate new ones for production!)
JWT_SECRET_KEY=<generate-with: openssl rand -base64 64>
INTERNAL_API_SECRET=<generate-with: openssl rand -base64 32>

# Admin Account
ADMIN_USERNAME=admin
ADMIN_EMAIL=maplesense2025@gmail.com
ADMIN_PASSWORD=<strong-password>

# MQTT Broker (if you want authentication)
BROKER_USER=<mqtt-username>
BROKER_PASS=<mqtt-password>
BROKER_TLS=false

# MQTT Topics
MQTT_TOPIC=sensors/#
MQTT_CLIENT_ID=mqtt-ingestor-prod

# OpenWeather API Key
OPENWEATHER_API_KEY=14ef204fbb7a18e0dda966c94fe7533b

# Email (Brevo SMTP) - for MQT EmailService alerts
SMTP_USERNAME=<brevo-smtp-login>
SMTP_PASSWORD=<brevo-smtp-key>
EMAIL_FROM_ADDRESS=<sender@yourdomain.com>
EMAIL_FROM_NAME=MapleSense Alerts
```

**Generate secrets:**

```bash
# Generate JWT secret
openssl rand -base64 64

# Generate internal API secret
openssl rand -base64 32
```

**Set file permissions:**

```bash
chmod 600 .env.production
```

---

## Step 4: Set Up Mosquitto Files (Optional but Recommended)

The `docker-compose.rpi.yml` expects these directories (they're optional if mosquitto allows anonymous):

```bash
# Create directories
mkdir -p mosquitto/certs mosquitto/config/passwd mosquitto/config/acl

# If you want MQTT authentication, create password file:
# (You'll need mosquitto_passwd tool installed)
mosquitto_passwd -c mosquitto/config/passwd <username>
# Enter password when prompted

# Create ACL file (mosquitto/config/acl) if needed:
# Example: user <username> topic readwrite sensors/#
```

**Note:** If you don't set up passwords/ACL, mosquitto will allow anonymous connections (as per the default config). For production, you should set up authentication.

---

## Step 5: Verify Docker Hub Images Exist

Make sure these images are pushed to Docker Hub (use `./push-to-dockerhub.sh` to build and push all):

- `maplesense/mptt-server-mosquitto:v1.0.0`
- `maplesense/mptt-server-api:v1.0.0`
- `maplesense/mptt-server-frontend:v1.0.0`
- `maplesense/mptt-server-bridge:v1.0.0`
- `maplesense/mptt-server-ingestor:v1.0.0`
- `maplesense/mptt-server-email:v1.0.0`

**Check on Docker Hub:** https://hub.docker.com/r/maplesense

---

## Step 6: Login to Docker Hub (if pulling private images)

```bash
docker login
# Enter your Docker Hub username and password
```

**Note:** If your images are public, you can skip this step.

---

## Step 7: Deploy the Stack

```bash
# Make deploy script executable
chmod +x deploy-rpi.sh

# Run deployment
./deploy-rpi.sh
```

**What this does:**

1. Validates `.env.production`
2. Pulls all images from Docker Hub using `DOCKERHUB_USERNAME` and `IMAGE_TAG`
3. Creates `mqtt-network` Docker network
4. Starts all services:
   - `mosquitto` (MQTT broker, internal only)
   - `postgres` (database)
   - `mqtt-bridge` (external broker bridge)
   - `mqtt-ingestor` (data ingestion)
   - `api-service` (REST API)
   - `mqt-frontend` (Next.js UI)
   - `nginx` (reverse proxy, port 80)
   - `cloudflared` (Cloudflare Tunnel connector)

---

## Step 8: Verify Deployment

### Check Running Containers

```bash
docker compose -f docker-compose.rpi.yml ps
```

You should see all containers in `running` state, especially:
- `nginx-proxy`
- `cloudflared`

### Check Logs

```bash
# All services
docker compose -f docker-compose.rpi.yml logs -f

# Specific service
docker compose -f docker-compose.rpi.yml logs -f cloudflared
docker compose -f docker-compose.rpi.yml logs -f nginx-proxy
```

### Test Locally on Pi

```bash
# Test nginx locally
curl http://localhost/health

# Should return API health status
```

---

## Step 9: Configure Cloudflare Tunnel (in Cloudflare Dashboard)

1. Go to **Cloudflare Zero Trust** → **Networks** → **Tunnels**
2. Find your tunnel (the one you created to get the token)
3. Go to **Public Hostnames** tab
4. Add hostname:
   - **Hostname:** `orpheus-networks.com`
   - **Service:** `http://nginx:80`
   - **Path:** (leave empty)
5. (Optional) Add `www.orpheus-networks.com` → `http://nginx:80`

**Important:** Cloudflare will automatically create DNS entries (usually proxied CNAMEs) for these hostnames.

---

## Step 10: Test External Access

From **outside your campus network** (mobile data or another network):

### Test Web UI

Open in browser:
- `https://orpheus-networks.com`

Should show your MQT frontend.

### Test API

```bash
curl https://orpheus-networks.com/api/health
```

Should return API health status.

### Test MQTT WebSocket (Optional)

Using a WebSocket client or MQTT.js:

```javascript
// Example using MQTT.js in Node.js
const mqtt = require('mqtt');

const client = mqtt.connect('wss://orpheus-networks.com/mqtt', {
  clientId: 'test-client'
});

client.on('connect', () => {
  console.log('Connected to MQTT via WebSocket!');
  client.subscribe('sensors/#');
});

client.on('message', (topic, message) => {
  console.log(`Received: ${topic} -> ${message.toString()}`);
});
```

---

## Troubleshooting

### Cloudflared Not Running

```bash
# Check logs
docker compose -f docker-compose.rpi.yml logs cloudflared

# Common issues:
# - Invalid TUNNEL_TOKEN
# - Network connectivity issues
# - Tunnel not configured in Cloudflare dashboard
```

### Nginx Not Responding

```bash
# Check nginx logs
docker compose -f docker-compose.rpi.yml logs nginx-proxy

# Test nginx config
docker exec nginx-proxy nginx -t

# Restart nginx
docker compose -f docker-compose.rpi.yml restart nginx
```

### Images Not Pulling

```bash
# Check Docker Hub login
docker login

# Manually pull an image to test
docker pull maplesense/mptt-server-api:v1.0.0

# Check .env.production has correct DOCKERHUB_USERNAME and IMAGE_TAG
```

### Services Can't Connect to Each Other

```bash
# Check Docker network
docker network inspect mqtt-network

# Verify all services are on the same network
docker compose -f docker-compose.rpi.yml ps
```

---

## Common Commands

```bash
# View all running containers
docker compose -f docker-compose.rpi.yml ps

# View logs (all services)
docker compose -f docker-compose.rpi.yml logs -f

# View logs (specific service)
docker compose -f docker-compose.rpi.yml logs -f api-service

# Restart a service
docker compose -f docker-compose.rpi.yml restart api-service

# Stop all services
docker compose -f docker-compose.rpi.yml down

# Start services again
docker compose -f docker-compose.rpi.yml up -d

# Pull latest images and restart
docker compose -f docker-compose.rpi.yml pull
docker compose -f docker-compose.rpi.yml up -d
```

---

## Summary Checklist

- [ ] Docker & Docker Compose installed on Pi
- [ ] Repository cloned on Pi
- [ ] `.env.production` created with all values (including `CLOUDFLARE_TUNNEL_TOKEN`)
- [ ] Mosquitto directories created (optional)
- [ ] Docker Hub images exist and are accessible
- [ ] `./deploy-rpi.sh` runs successfully
- [ ] All containers running (`docker compose ps`)
- [ ] Cloudflare Tunnel configured with hostname `orpheus-networks.com` → `http://nginx:80`
- [ ] External access works: `https://orpheus-networks.com`
- [ ] API works: `https://orpheus-networks.com/api/health`
- [ ] MQTT WebSocket works: `wss://orpheus-networks.com/mqtt`

---

## Next Steps

- Monitor logs: `docker compose -f docker-compose.rpi.yml logs -f`
- Set up monitoring/alerting for services
- Configure automatic restarts (already set with `restart: unless-stopped`)
- Consider setting up backups for PostgreSQL data volume
