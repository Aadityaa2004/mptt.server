FROM nginx:alpine

# Install certbot and openssl for certificate management
RUN apk add --no-cache certbot certbot-nginx openssl

# Copy Nginx configuration files
COPY nginx.conf /etc/nginx/nginx.conf
COPY default.conf /etc/nginx/conf.d/default.conf

# Create directories for Let's Encrypt certificates and webroot
RUN mkdir -p /var/www/certbot /etc/letsencrypt

# Expose ports
EXPOSE 80 443

# Start script that handles certificate renewal
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]

