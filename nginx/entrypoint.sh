#!/bin/sh

# Create webroot directory if it doesn't exist
mkdir -p /var/www/certbot

# Start Nginx in background for initial certificate generation
nginx -g "daemon on;"

# Wait for Nginx to start
sleep 2

# Check if certificates exist
if [ ! -f /etc/letsencrypt/live/orpheus-networks.com/fullchain.pem ]; then
    echo "No SSL certificates found. Please run certbot to generate certificates."
    echo "You can run: certbot certonly --webroot -w /var/www/certbot -d orpheus-networks.com -d www.orpheus-networks.com"
else
    echo "SSL certificates found. Starting Nginx..."
fi

# Stop background Nginx
nginx -s quit

# Wait for Nginx to stop
sleep 2

# Start Nginx in foreground
exec nginx -g "daemon off;"

