# Raspberry Pi End-to-End Deployment (orpheus-networks.com)

This guide walks you from **zero to live** on a Raspberry Pi, using `orpheus-networks.com` with the integrated nginx setup.

It assumes you will:
- Build and **push versioned Docker images** to Docker Hub
- Deploy using `docker-compose.rpi.yml` (which includes nginx)
- Start with **HTTP (port 80)**, then optionally add **HTTPS (port 443)** via Let's Encrypt

---

## 1. Find your public IP address

From any machine on the same network as the Pi:

```bash
curl ifconfig.me
# or open https://whatismyipaddress.com in a browser
```

Note the IPv4 address (e.g. `203.0.113.42`). This is what you should put into GoDaddy, **not** your internal 172.x / 192.168.x address.

---

## 2. Configure DNS in GoDaddy

In GoDaddy DNS management for `orpheus-networks.com`:

- **A record for root domain**
  - **Name:** `@`
  - **Type:** A
  - **Value:** your **public IP** from step 1
  - **TTL:** 1 hour (or default)

- **A record for www**
  - **Name:** `www`
  - **Type:** A
  - **Value:** same **public IP**
  - **TTL:** 1 hour

Wait for DNS to propagate (often a few minutes, up to a couple of hours).

Optional verification:

```bash
dig orpheus-networks.com
nslookup orpheus-networks.com
```

You should see the public IP from step 1.

---

## 3. Router / network notes

For **home router deployments**, you can still port-forward 80/443/8883/9443 directly to the Pi as described in `DEPLOYMENT.md`.

For **UMass campus ethernet or any network where you cannot port-forward**, you should:

- Use **Cloudflare** as DNS provider for `orpheus-networks.com`
- Use **Cloudflare Tunnel** (via the `cloudflared` service in `docker-compose.rpi.yml`) instead of exposing ports on the campus network

In the campus case, you do **not** open inbound ports on the router; Cloudflare connects _out_ from the Pi to Cloudflare and reverses traffic through the tunnel.

---

## 4. Build & push versioned Docker images

You will push images like:

- `${DOCKERHUB_USERNAME}/mptt-server-mosquitto:v1.0.0`
- `${DOCKERHUB_USERNAME}/mptt-server-api:v1.0.0`
- `${DOCKERHUB_USERNAME}/mptt-server-frontend:v1.0.0`
- `${DOCKERHUB_USERNAME}/mptt-server-bridge:v1.0.0`
- `${DOCKERHUB_USERNAME}/mptt-server-ingestor:v1.0.0`

### 4.1 Choose a version

Pick a version string, e.g.:

```bash
export VERSION=v1.0.0
```

You’ll reuse this both when **pushing** and when **deploying**.

### 4.2 Push images using the script

From your dev machine in the repo root:

```bash
chmod +x push-to-dockerhub.sh

# Replace YOUR_DOCKERHUB_USER with your Docker Hub username
VERSION=$VERSION ./push-to-dockerhub.sh YOUR_DOCKERHUB_USER
```

The script will:
- Build and push all services with tag `$VERSION`
- Also push `:latest` for convenience

You can confirm on Docker Hub that the `v1.0.0` tags exist for each image.

For ARM64-specific or multi-arch builds, see `BUILD_ARM64.md`.

---

## 5. Prepare `.env.production` on the Pi

On the **Pi**, clone the repo and create the env file:

```bash
git clone <your-repo-url>
cd mptt.server

cp env.production.example .env.production
nano .env.production
```

Set at least:

- **Domain / URLs**
  - `DOMAIN=orpheus-networks.com`
  - `NEXT_PUBLIC_API_BASE_URL=https://orpheus-networks.com/api` (for final HTTPS)
  - `NEXT_PUBLIC_READINGS_API_BASE_URL=https://orpheus-networks.com/api`
  - `CORS_ALLOWED_ORIGINS=https://orpheus-networks.com,https://www.orpheus-networks.com`

- **Docker Hub / image tag**
  - `DOCKERHUB_USERNAME=YOUR_DOCKERHUB_USER`
  - `IMAGE_TAG=v1.0.0`  (must match the VERSION you used when pushing)

- **Postgres / secrets / admin**
  - `POSTGRES_PASSWORD=...`
  - `JWT_SECRET_KEY=...` (generate with `openssl rand -base64 64`)
  - `INTERNAL_API_SECRET=...` (generate with `openssl rand -base64 32`)
  - `ADMIN_PASSWORD=...`
  - `BROKER_USER`, `BROKER_PASS` for MQTT

Then lock down permissions:

```bash
chmod 600 .env.production
```

> Note: `.env.production` is used both to configure services and to provide `DOCKERHUB_USERNAME` / `IMAGE_TAG` for `docker-compose.rpi.yml`.

---

## 6. Mosquitto TLS files (MQTT)

The RPi compose mounts mosquitto certs:

```yaml
mosquitto:
  volumes:
    - ./mosq-data:/mosquitto/data
    - ./mosq-logs:/mosquitto/log
    - ./mosquitto/certs:/mosquitto/certs:ro
    - ./mosquitto/config/passwd:/mosquitto/config/passwd:ro
    - ./mosquitto/config/acl:/mosquitto/config/acl:ro
```

At minimum:

- Place your **broker certificate and key** (and CA) under `mosquitto/certs` as expected by the mosquitto config in `mosquitto/config`.
- Set up your MQTT username/password in `mosquitto/config/passwd` and reference them via `BROKER_USER` / `BROKER_PASS` in `.env.production`.

For detailed mosquitto TLS setup, refer to the mosquitto docs and any notes in `DEPLOYMENT.md`.

---

## 7. Deploy the stack on the Pi (HTTP via nginx)

From the repo root on the Pi:

```bash
./deploy-rpi.sh
```

This will:

1. Validate `.env.production`
2. Pull images from Docker Hub using `DOCKERHUB_USERNAME` and `IMAGE_TAG`
3. Create the `mqtt-network`
4. Start all services from `docker-compose.rpi.yml`, including:
   - `mosquitto`
   - `postgres`
   - `mqtt-bridge`
   - `mqtt-ingestor`
   - `api-service`
   - `mqt-frontend`
   - `nginx` (HTTP-only)

Confirm containers:

```bash
docker ps
```

At this point, with DNS and either router port forwarding (home) or Cloudflare Tunnel (campus) correct:

- `http://orpheus-networks.com/` → MQT frontend
- `http://orpheus-networks.com/api/health` → API health endpoint

If you are still on HTTP-only, the frontend is configured for HTTPS API, so API calls may fail until you add SSL. You can temporarily use `http://orpheus-networks.com:3000` for testing, or proceed to SSL.

---

## 8. Obtain SSL certificates (nginx + Let’s Encrypt)

Prerequisites:

- DNS A records for `orpheus-networks.com` and `www.orpheus-networks.com` point to your **public IP**
- Router forwards **80** and **443** to the Pi
- `nginx-proxy` container is running (from `deploy-rpi.sh`)

Run on the Pi:

```bash
./setup-ssl.sh
```

The script will:

1. Ensure `nginx-proxy` is running
2. Run `certbot` inside the container using the webroot at `/var/www/certbot`
3. Request certificates for `DOMAIN` and `www.DOMAIN`
4. Set up a `renew-certs.sh` script and cron job for automatic renewal

You can re-run the script later to renew or repair certificates.

---

## 9. Switch nginx to HTTPS (optional, recommended)

Currently, `docker-compose.rpi.yml` mounts the **HTTP-only** config:

```yaml
nginx:
  volumes:
    - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
    - ./nginx/default-http-only.conf:/etc/nginx/conf.d/default.conf:ro
```

After certificates exist (via `setup-ssl.sh`), you can switch to the full HTTPS config:

1. Edit `docker-compose.rpi.yml`:
   - Change the `default-http-only.conf` mount to `nginx/default.conf`
   - Uncomment/add the `443:443` port mapping

2. Redeploy nginx:

```bash
docker compose -f docker-compose.rpi.yml up -d nginx
```

3. Test:

```bash
curl https://orpheus-networks.com/api/health
```

In the browser, open:

- `https://orpheus-networks.com`

Your existing `.env.production` values for `NEXT_PUBLIC_API_BASE_URL` and `CORS_ALLOWED_ORIGINS` are already set up for HTTPS, so the frontend and API should align once SSL is active.

---

## 10. Quick checklist

- **DNS**
  - [ ] `@` and `www` A records → public IP
- **Router**
  - [ ] 80 → Pi:80, 443 → Pi:443, 8883/9443 → Pi
- **Docker Hub**
  - [ ] Images pushed with `IMAGE_TAG` (e.g. `v1.0.0`)
- **Pi env**
  - [ ] `.env.production` with `DOCKERHUB_USERNAME` and `IMAGE_TAG`
- **Deployment**
  - [ ] `./deploy-rpi.sh` runs successfully
- **HTTP**
  - [ ] `http://orpheus-networks.com` shows the UI
- **SSL (optional but recommended)**
  - [ ] `./setup-ssl.sh` completed
  - [ ] nginx switched to `nginx/default.conf` and port 443 exposed
  - [ ] `https://orpheus-networks.com` works

---

## 11. Campus / Cloudflare Tunnel deployment (no port forwarding)

If your Pi is on a network where you **cannot port-forward** (e.g. UMass campus ethernet), use this Cloudflare Tunnel flow instead of exposing ports directly:

### 11.1 Move DNS to Cloudflare

1. Create a Cloudflare account and add the domain `orpheus-networks.com`.
2. In your current registrar (e.g. GoDaddy), set the domain’s nameservers to the ones given by Cloudflare.
3. In Cloudflare DNS:
   - Make sure there are **no A-records pointing to a campus IP**.
   - When you create a tunnel hostname (below), Cloudflare will manage the DNS for you (usually a proxied CNAME).

### 11.2 Create a Cloudflare Tunnel (in Cloudflare dashboard)

1. In Cloudflare, go to **Zero Trust** → **Networks** → **Tunnels**.
2. Create a new tunnel (e.g. `orpheus-rpi-tunnel`).
3. Choose **Docker** as the connector option and copy the **Tunnel token** (it looks like a long string used as `TUNNEL_TOKEN`).
4. Under the tunnel’s **Public Hostnames**, add:
   - **Hostname:** `orpheus-networks.com`
   - **Service:** `http://nginx:80`
   - (Optionally) `www.orpheus-networks.com` → `http://nginx:80` as well.

Cloudflare will create the necessary DNS entries for you.

### 11.3 Configure tunnel token on the Pi

On the Pi, in `.env.production`, set:

```bash
CLOUDFLARE_TUNNEL_TOKEN=<paste-from-cloudflare>
```

Leave your `DOCKERHUB_USERNAME`, `IMAGE_TAG`, and other settings as before.

The `docker-compose.rpi.yml` file includes a `cloudflared` service:

```259:272:/Users/aadityaa/Documents/Codebase/SDP/mptt.server/docker-compose.rpi.yml
cloudflared:
  image: cloudflare/cloudflared:latest
  container_name: cloudflared
  restart: unless-stopped
  depends_on:
    - nginx
  environment:
    - TUNNEL_TOKEN=${CLOUDFLARE_TUNNEL_TOKEN}
  command: tunnel --no-autoupdate run
  networks:
    - mqtt-network
```

This container dials out to Cloudflare and registers the tunnel.

### 11.4 Deploy on campus

From the repo root on the Pi:

```bash
./deploy-rpi.sh
```

This will:

- Pull all images from Docker Hub
- Start:
  - `mosquitto` (MQTT broker)
  - `postgres`
  - `mqtt-bridge`
  - `mqtt-ingestor`
  - `api-service`
  - `mqt-frontend`
  - `nginx` (HTTP only, internal container)
  - `cloudflared` (Cloudflare Tunnel connector)

Check containers:

```bash
docker compose -f docker-compose.rpi.yml ps
```

You should see `cloudflared` in a `running` state.

### 11.5 Access through Cloudflare

From any external network (off-campus or mobile data):

- **Web UI:** `https://orpheus-networks.com`
- **API:** `https://orpheus-networks.com/api`
- **MQTT WebSocket:** `wss://orpheus-networks.com/mqtt`

Internally, nginx handles:

- `/` → `http://mqt-frontend:3000`
- `/api` → `http://api-service:9002`
- `/mqtt` → `http://mosquitto:9001` (websockets)

These all run over HTTP inside the Docker network, while Cloudflare provides TLS at the edge for external clients.

