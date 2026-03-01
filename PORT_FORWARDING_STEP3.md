# Step 3: Router Port Forwarding Setup

This guide explains how to configure your router to forward incoming internet traffic to your Raspberry Pi, allowing external access to your services.

## Overview

Port forwarding tells your router to send specific incoming internet traffic to your Raspberry Pi. Without this, your Raspberry Pi is only accessible on your local network.

## Prerequisites

1. **Router Admin Access**: Login credentials for your router's admin panel
2. **Raspberry Pi Local IP**: The IP address of your Raspberry Pi on your local network
3. **Router Model**: Know your router brand/model (helps with specific instructions)

## Required Ports

You need to forward these ports to your Raspberry Pi:

| Port | Protocol | Service | Description |
|------|----------|---------|-------------|
| 80 | TCP | HTTP | Nginx (for Let's Encrypt SSL certificate) |
| 443 | TCP | HTTPS | Nginx (SSL/TLS encrypted web traffic) |
| 8883 | TCP | MQTT TLS | MQTT Broker (for IoT devices) |
| 9443 | TCP | MQTT WebSocket TLS | MQTT WebSocket (for web clients) |

**Note:** Port 22 (SSH) is optional but recommended for remote management.

## Step-by-Step Instructions

### Step 1: Find Your Raspberry Pi's Local IP Address

On your Raspberry Pi, run:

```bash
hostname -I
# or
ip addr show
# or
ifconfig
```

Look for the IP address that starts with:
- `192.168.x.x` (most common)
- `10.x.x.x`
- `172.16.x.x` to `172.31.x.x`

**Example:** `192.168.1.100`

**Note this IP address** - you'll need it for port forwarding configuration.

### Step 2: Access Your Router's Admin Panel

#### Find Your Router's IP Address

**On Windows:**
```cmd
ipconfig
# Look for "Default Gateway"
```

**On macOS/Linux:**
```bash
route -n get default | grep gateway
# or
netstat -rn | grep default
```

**Common Router IPs:**
- `192.168.1.1` (most common)
- `192.168.0.1`
- `10.0.0.1`
- `192.168.2.1`

#### Access Router Admin

1. Open a web browser
2. Navigate to your router's IP (e.g., `http://192.168.1.1`)
3. Log in with admin credentials

**Common Default Credentials:**
- Username: `admin` / Password: `admin`
- Username: `admin` / Password: `password`
- Username: `admin` / Password: (blank)
- Check router label or manual for default credentials

**If you changed the password and forgot it:**
- Reset router to factory defaults (button on back)
- Or check router documentation

### Step 3: Locate Port Forwarding Settings

The location varies by router brand. Look for:

- **Port Forwarding**
- **Virtual Server**
- **Port Mapping**
- **NAT Forwarding**
- **Applications & Gaming**
- **Firewall Rules**
- **Advanced → Port Forwarding**

### Step 4: Configure Port Forwarding Rules

Create a port forwarding rule for each port. Here's what each rule needs:

#### Rule 1: HTTP (Port 80)

- **Service Name/Description:** `HTTP` or `Web Server`
- **External Port:** `80`
- **Internal Port:** `80`
- **Protocol:** `TCP` (or `Both`/`TCP/UDP`)
- **Internal IP:** Your Raspberry Pi's IP (e.g., `192.168.1.100`)
- **Enable:** `Yes` or checked

#### Rule 2: HTTPS (Port 443)

- **Service Name/Description:** `HTTPS` or `SSL`
- **External Port:** `443`
- **Internal Port:** `443`
- **Protocol:** `TCP`
- **Internal IP:** Your Raspberry Pi's IP
- **Enable:** `Yes`

#### Rule 3: MQTT TLS (Port 8883)

- **Service Name/Description:** `MQTT TLS`
- **External Port:** `8883`
- **Internal Port:** `8883`
- **Protocol:** `TCP`
- **Internal IP:** Your Raspberry Pi's IP
- **Enable:** `Yes`

#### Rule 4: MQTT WebSocket TLS (Port 9443)

- **Service Name/Description:** `MQTT WebSocket`
- **External Port:** `9443`
- **Internal Port:** `9443`
- **Protocol:** `TCP`
- **Internal IP:** Your Raspberry Pi's IP
- **Enable:** `Yes`

#### Optional: SSH (Port 22)

- **Service Name/Description:** `SSH`
- **External Port:** `22`
- **Internal Port:** `22`
- **Protocol:** `TCP`
- **Internal IP:** Your Raspberry Pi's IP
- **Enable:** `Yes`

**Security Note:** Only enable SSH if you need remote access. Consider changing the default SSH port for better security.

### Step 5: Save and Apply Changes

1. **Save** all port forwarding rules
2. **Apply** changes (router may restart)
3. **Wait** for router to finish restarting (1-2 minutes)

### Step 6: Verify Port Forwarding

#### Method 1: Test from External Network

1. **Disconnect from your WiFi** (use mobile data or different network)
2. **Test HTTP access:**
   ```bash
   curl http://YOUR_PUBLIC_IP
   # or visit http://YOUR_PUBLIC_IP in browser
   ```
3. **Test HTTPS access:**
   ```bash
   curl https://YOUR_PUBLIC_IP
   # or visit https://YOUR_PUBLIC_IP in browser
   ```

#### Method 2: Online Port Checkers

Use online tools to check if ports are open:

- https://www.yougetsignal.com/tools/open-ports/
- https://canyouseeme.org
- https://www.portchecker.co

Enter your public IP and check ports: 80, 443, 8883, 9443

**What to expect:**
- ✅ **Open/Reachable** - Port forwarding is working
- ❌ **Closed/Filtered** - Port forwarding not configured or blocked

#### Method 3: Test from Raspberry Pi

On your Raspberry Pi, check if services are listening:

```bash
# Check if services are running on ports
sudo netstat -tuln | grep -E ':(80|443|8883|9443)'
# or
sudo ss -tuln | grep -E ':(80|443|8883|9443)'
```

## Router-Specific Instructions

### Netgear Routers

1. Log into router admin (usually `192.168.1.1`)
2. Go to **Advanced** → **Advanced Setup** → **Port Forwarding / Port Triggering**
3. Click **Add Custom Service**
4. Fill in:
   - **Service Name:** (e.g., "HTTP")
   - **External Starting Port:** 80
   - **External Ending Port:** 80
   - **Internal Starting Port:** 80
   - **Internal Ending Port:** 80
   - **Internal IP Address:** Your Raspberry Pi IP
5. Click **Apply**
6. Repeat for other ports

### TP-Link Routers

1. Log into router admin (usually `192.168.0.1` or `192.168.1.1`)
2. Go to **Advanced** → **NAT Forwarding** → **Virtual Servers**
3. Click **Add**
4. Fill in:
   - **Service Type:** Custom
   - **External Port:** 80
   - **Internal Port:** 80
   - **Internal IP:** Your Raspberry Pi IP
   - **Protocol:** TCP
   - **Status:** Enabled
5. Click **Save**
6. Repeat for other ports

### ASUS Routers

1. Log into router admin (usually `192.168.1.1`)
2. Go to **WAN** → **Virtual Server / Port Forwarding**
3. Click **Add Profile**
4. Fill in:
   - **Service Name:** HTTP
   - **Port Range:** 80
   - **Local IP:** Your Raspberry Pi IP
   - **Local Port:** 80
   - **Protocol:** TCP
5. Click **OK**
6. Repeat for other ports

### Linksys Routers

1. Log into router admin (usually `192.168.1.1`)
2. Go to **Connectivity** → **Router Settings** → **Port Forwarding**
3. Click **Add**
4. Fill in:
   - **Application Name:** HTTP
   - **External Port:** 80
   - **Internal Port:** 80
   - **Protocol:** TCP
   - **Device IP:** Your Raspberry Pi IP
5. Click **OK**
6. Repeat for other ports

### D-Link Routers

1. Log into router admin (usually `192.168.0.1`)
2. Go to **Advanced** → **Port Forwarding**
3. Click **Add Rule**
4. Fill in:
   - **Rule Name:** HTTP
   - **External Port Start:** 80
   - **External Port End:** 80
   - **Internal Port Start:** 80
   - **Internal Port End:** 80
   - **Internal IP:** Your Raspberry Pi IP
   - **Protocol:** TCP
5. Click **Save**
6. Repeat for other ports

### Google Nest WiFi / Google WiFi

1. Open Google Home app
2. Select your WiFi network
3. Go to **Settings** → **Advanced networking** → **Port management**
4. Click **Add** → **Port forwarding**
5. Fill in:
   - **Name:** HTTP
   - **Internal IP:** Your Raspberry Pi IP
   - **External Port:** 80
   - **Internal Port:** 80
   - **Protocol:** TCP
6. Click **Save**
7. Repeat for other ports

### Ubiquiti UniFi

1. Log into UniFi Controller
2. Go to **Settings** → **Networks** → Select your network
3. Scroll to **Port Forwarding** section
4. Click **Add Port Forwarding Rule**
5. Fill in:
   - **Name:** HTTP
   - **From:** WAN
   - **To:** Your Raspberry Pi IP
   - **Port:** 80
   - **Protocol:** TCP
6. Click **Apply Changes**
7. Repeat for other ports

## Troubleshooting

### Ports Show as Closed

**Problem:** Online port checker shows ports as closed

**Solutions:**
1. **Verify Raspberry Pi IP is correct** - Check IP hasn't changed
2. **Check services are running** - Services must be running on Raspberry Pi
3. **Check router firewall** - Some routers have firewall that blocks ports
4. **Verify port forwarding rules** - Double-check all settings
5. **Restart router** - Sometimes rules need a restart to take effect
6. **Check ISP restrictions** - Some ISPs block certain ports (80, 443 may be blocked)

### Can't Access Router Admin

**Problem:** Can't log into router admin panel

**Solutions:**
1. **Try different IPs:**
   - `192.168.1.1`
   - `192.168.0.1`
   - `10.0.0.1`
   - `192.168.2.1`
2. **Check router label** - May have admin IP printed
3. **Reset router** - Factory reset if needed (loses all settings)
4. **Check connection** - Make sure you're connected to router's network

### Port Forwarding Not Working

**Problem:** Configured port forwarding but still can't access from internet

**Solutions:**
1. **Check Raspberry Pi firewall:**
   ```bash
   # On Raspberry Pi
   sudo ufw status
   # If enabled, allow ports:
   sudo ufw allow 80/tcp
   sudo ufw allow 443/tcp
   sudo ufw allow 8883/tcp
   sudo ufw allow 9443/tcp
   ```

2. **Verify services are running:**
   ```bash
   # Check if services are listening
   sudo netstat -tuln | grep -E ':(80|443|8883|9443)'
   ```

3. **Test locally first:**
   ```bash
   # On Raspberry Pi, test if services respond locally
   curl http://localhost:80
   curl http://localhost:443
   ```

4. **Check router logs** - Some routers show connection attempts in logs

5. **Try different external port** - Some ISPs block common ports

### ISP Blocks Ports 80 and 443

**Problem:** ISP blocks HTTP/HTTPS ports (common on residential connections)

**Solutions:**
1. **Use different ports:**
   - Forward external port 8080 → internal port 80
   - Forward external port 8443 → internal port 443
   - Access via: `http://your-domain.com:8080`

2. **Contact ISP** - Some ISPs can unblock ports for business accounts

3. **Use reverse proxy service** - Services like Cloudflare Tunnel or ngrok

### Dynamic IP Address

**Problem:** Raspberry Pi IP changes, breaking port forwarding

**Solutions:**
1. **Set static IP on Raspberry Pi:**
   ```bash
   # Edit network config
   sudo nano /etc/dhcpcd.conf
   
   # Add at end:
   interface eth0  # or wlan0 for WiFi
   static ip_address=192.168.1.100/24
   static routers=192.168.1.1
   static domain_name_servers=192.168.1.1 8.8.8.8
   
   # Restart networking
   sudo systemctl restart dhcpcd
   ```

2. **Set static IP in router:**
   - Router admin → DHCP settings
   - Reserve IP for Raspberry Pi's MAC address

## Security Considerations

### Firewall on Raspberry Pi

Enable firewall and only allow necessary ports:

```bash
# Install UFW if not installed
sudo apt install ufw

# Allow SSH (if needed)
sudo ufw allow 22/tcp

# Allow HTTP/HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Allow MQTT
sudo ufw allow 8883/tcp
sudo ufw allow 9443/tcp

# Enable firewall
sudo ufw enable

# Check status
sudo ufw status
```

### Change Default SSH Port (Optional)

For better security, change SSH port:

```bash
# Edit SSH config
sudo nano /etc/ssh/sshd_config

# Change:
Port 22
# To:
Port 2222  # or any port you prefer

# Restart SSH
sudo systemctl restart sshd

# Update port forwarding rule in router to new port
```

### Regular Security Updates

Keep Raspberry Pi updated:

```bash
sudo apt update && sudo apt upgrade -y
```

## Testing Your Setup

### Complete Test Checklist

1. **Local Network Test:**
   ```bash
   # From another device on same network
   curl http://RASPBERRY_PI_LOCAL_IP:80
   ```

2. **External Network Test:**
   ```bash
   # From different network (mobile data)
   curl http://YOUR_PUBLIC_IP:80
   curl https://YOUR_PUBLIC_IP:443
   ```

3. **Domain Test (after DNS is set up):**
   ```bash
   curl http://orpheus-networks.com
   curl https://orpheus-networks.com
   ```

4. **Port Checker Test:**
   - Use online port checker tools
   - Verify ports 80, 443, 8883, 9443 are open

## Quick Checklist

- [ ] Found Raspberry Pi local IP address
- [ ] Accessed router admin panel
- [ ] Located port forwarding settings
- [ ] Created port forwarding rule for port 80
- [ ] Created port forwarding rule for port 443
- [ ] Created port forwarding rule for port 8883
- [ ] Created port forwarding rule for port 9443
- [ ] Saved and applied changes
- [ ] Verified ports are open (using online checker)
- [ ] Tested access from external network
- [ ] Configured Raspberry Pi firewall (if needed)

## Next Steps

After port forwarding is configured:

1. ✅ **Step 3 Complete** - Ports are forwarded to Raspberry Pi
2. **Step 4** - Deploy services on Raspberry Pi
3. **Step 5** - Set up SSL certificates
4. **Step 6** - Verify everything is working

## Important Notes

- **Keep Raspberry Pi IP static** - Use DHCP reservation or static IP configuration
- **Test from external network** - Port forwarding only works from outside your network
- **Some ISPs block ports** - Ports 80/443 may be blocked on residential connections
- **Security matters** - Enable firewall and keep system updated
- **Document your settings** - Write down your port forwarding rules for reference

Once port forwarding is working, you're ready to deploy your services!

