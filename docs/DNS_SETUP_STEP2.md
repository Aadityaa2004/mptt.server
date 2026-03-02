# Step 2: Configure DNS for orpheus-networks.com

This guide explains how to configure DNS records so that `orpheus-networks.com` points to your Raspberry Pi's public IP address.

## Overview

DNS (Domain Name System) translates domain names like `orpheus-networks.com` into IP addresses. You need to configure DNS records so that when someone visits your domain, they're directed to your Raspberry Pi's public IP address.

## Prerequisites

1. **Domain Name**: You own `orpheus-networks.com` (or have access to manage its DNS)
2. **Public IP Address**: Your Raspberry Pi's public IP address (from your internet service provider)
3. **Access to DNS Management**: Login credentials for your domain registrar or DNS provider

## Step-by-Step Instructions

### Step 1: Find Your Public IP Address

Your Raspberry Pi needs a public IP address that the internet can reach. There are two scenarios:

#### Scenario A: Static Public IP (Recommended)

If your internet service provider (ISP) gave you a static IP address, use that.

#### Scenario B: Dynamic Public IP (Most Common)

If your IP changes periodically, you'll need to:

1. **Find your current public IP:**
   ```bash
   # On Raspberry Pi or any device on your network
   curl ifconfig.me
   # or
   curl ipinfo.io/ip
   # or visit: https://whatismyipaddress.com
   ```

2. **Note down this IP address** - you'll need it for DNS configuration

3. **Consider Dynamic DNS (Optional but Recommended):**
   - If your IP changes frequently, consider using a Dynamic DNS service
   - Services like DuckDNS, No-IP, or DynDNS can automatically update DNS when your IP changes
   - This is optional but makes management easier

### Step 2: Access Your DNS Management Panel

Where you manage DNS depends on where you registered your domain:

**Common Domain Registrars:**
- GoDaddy
- Namecheap
- Google Domains
- Cloudflare
- Name.com
- Hover
- AWS Route 53
- DigitalOcean

**Steps:**
1. Log into your domain registrar's website
2. Navigate to "DNS Management", "DNS Settings", or "Domain Management"
3. Look for "DNS Records", "DNS Zone", or "Name Servers"

### Step 3: Configure DNS Records

You need to create **A Records** that point your domain to your public IP address.

#### Required DNS Records

Create these two A records:

**Record 1: Root Domain**
- **Type:** A
- **Name/Host:** `@` or `orpheus-networks.com` (or leave blank)
- **Value/Points to:** `YOUR_PUBLIC_IP` (the IP you found in Step 1)
- **TTL:** 3600 (or default/automatic)

**Record 2: WWW Subdomain**
- **Type:** A
- **Name/Host:** `www`
- **Value/Points to:** `YOUR_PUBLIC_IP` (same IP as above)
- **TTL:** 3600 (or default/automatic)

#### Example Configuration

If your public IP is `203.0.113.45`, your DNS records should look like:

| Type | Name | Value | TTL |
|------|------|-------|-----|
| A | @ | 203.0.113.45 | 3600 |
| A | www | 203.0.113.45 | 3600 |

**Note:** Some DNS providers use different formats:
- Some use `@` for the root domain
- Some use a blank/empty name field
- Some require the full domain name `orpheus-networks.com`

### Step 4: Save and Wait for Propagation

1. **Save the DNS records** in your DNS management panel
2. **Wait for DNS propagation** - This can take:
   - **Minimum:** 5-15 minutes
   - **Typical:** 1-4 hours
   - **Maximum:** Up to 48 hours (rare)

### Step 5: Verify DNS Configuration

After saving, verify that DNS is working:

#### Method 1: Using `dig` (Linux/Mac)

```bash
# Check root domain
dig orpheus-networks.com +short

# Check www subdomain
dig www.orpheus-networks.com +short

# Both should return your public IP address
```

#### Method 2: Using `nslookup` (Windows/Linux/Mac)

```bash
# Check root domain
nslookup orpheus-networks.com

# Check www subdomain
nslookup www.orpheus-networks.com
```

#### Method 3: Online Tools

Visit these websites and enter your domain:
- https://dnschecker.org
- https://www.whatsmydns.net
- https://mxtoolbox.com/DNSLookup.aspx

**What to look for:**
- Both `orpheus-networks.com` and `www.orpheus-networks.com` should resolve to your public IP
- The IP address should match what you configured

## Common DNS Provider Guides

### Cloudflare

1. Log into Cloudflare dashboard
2. Select your domain `orpheus-networks.com`
3. Go to **DNS** → **Records**
4. Click **Add record**
5. **Type:** A
6. **Name:** `@` (for root) or `www` (for www subdomain)
7. **IPv4 address:** Your public IP
8. **Proxy status:** DNS only (gray cloud) - **Important:** Don't use proxy (orange cloud) initially
9. Click **Save**
10. Repeat for www subdomain

### GoDaddy

1. Log into GoDaddy account
2. Go to **My Products** → **Domains**
3. Click **DNS** next to your domain
4. Scroll to **Records** section
5. Click **Add** button
6. **Type:** A
7. **Name:** `@` (for root) or `www` (for www)
8. **Value:** Your public IP
9. **TTL:** 600 seconds (or default)
10. Click **Save**
11. Repeat for www subdomain

### Namecheap

1. Log into Namecheap account
2. Go to **Domain List**
3. Click **Manage** next to your domain
4. Go to **Advanced DNS** tab
5. Under **Host Records**, click **Add New Record**
6. **Type:** A Record
7. **Host:** `@` (for root) or `www` (for www)
8. **Value:** Your public IP
9. **TTL:** Automatic
10. Click **Save** (checkmark icon)
11. Repeat for www subdomain

### Google Domains

1. Log into Google Domains
2. Select your domain
3. Go to **DNS** section
4. Scroll to **Custom resource records**
5. Click **Add record**
6. **Name:** `@` (for root) or `www` (for www)
7. **Type:** A
8. **Data:** Your public IP
9. **TTL:** 3600
10. Click **Save**
11. Repeat for www subdomain

## Troubleshooting

### DNS Not Resolving

**Problem:** `dig` or `nslookup` doesn't return your IP

**Solutions:**
1. **Wait longer** - DNS propagation can take time
2. **Check for typos** - Verify the IP address is correct
3. **Clear DNS cache:**
   ```bash
   # On Linux
   sudo systemd-resolve --flush-caches
   
   # On macOS
   sudo dscacheutil -flushcache; sudo killall -HUP mDNSResponder
   
   # On Windows
   ipconfig /flushdns
   ```
4. **Check from different location** - Use online DNS checker tools
5. **Verify DNS records** - Double-check in your DNS management panel

### Wrong IP Address

**Problem:** DNS resolves but to wrong IP

**Solutions:**
1. Update the A record with correct IP
2. Wait for propagation
3. Clear DNS cache

### Dynamic IP Changes

**Problem:** Your public IP changes periodically

**Solutions:**
1. **Use Dynamic DNS service:**
   - Set up DuckDNS (free): https://www.duckdns.org
   - Set up No-IP (free): https://www.noip.com
   - Configure automatic IP updates
   - Point your domain to the Dynamic DNS hostname

2. **Update DNS manually** when IP changes:
   - Log into DNS management
   - Update A records with new IP
   - Wait for propagation

### Port Forwarding Not Working

**Note:** DNS configuration alone isn't enough. You also need:
- Router port forwarding (Step 3)
- Firewall configuration
- Services running on Raspberry Pi

## Testing Your Setup

Once DNS is configured and propagated:

1. **Test DNS resolution:**
   ```bash
   ping orpheus-networks.com
   # Should ping your public IP
   ```

2. **Test HTTP access (before SSL):**
   ```bash
   curl http://orpheus-networks.com
   # Should connect (may get SSL error, that's OK)
   ```

3. **Verify from browser:**
   - Visit `http://orpheus-networks.com`
   - Should connect to your Raspberry Pi (may see error page if services aren't running yet)

## Next Steps

After DNS is configured and propagated:

1. ✅ **Step 2 Complete** - DNS is pointing to your IP
2. **Step 3** - Set up router port forwarding (ports 80, 443, 8883, 9443)
3. **Step 4** - Deploy services on Raspberry Pi
4. **Step 5** - Set up SSL certificates

## Important Notes

- **DNS propagation takes time** - Be patient, it can take up to 48 hours (usually much faster)
- **Keep your public IP handy** - You'll need it for router configuration too
- **Test from multiple locations** - Use online DNS checkers to verify globally
- **Don't use DNS proxy initially** - If using Cloudflare, use "DNS only" mode (gray cloud) until SSL is set up

## Quick Checklist

- [ ] Found public IP address
- [ ] Logged into DNS management panel
- [ ] Created A record for `@` (root domain)
- [ ] Created A record for `www` subdomain
- [ ] Saved DNS records
- [ ] Verified DNS resolution (using dig/nslookup/online tools)
- [ ] Waited for propagation (if needed)

Once DNS is working, proceed to Step 3: Router Port Forwarding!

