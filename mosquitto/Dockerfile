# Mosquitto MQTT Broker Dockerfile
FROM eclipse-mosquitto:2

# Create necessary directories
RUN mkdir -p /mosquitto/data /mosquitto/log /mosquitto/config /mosquitto/certs

# Copy configuration files
# For development, use the simple config
COPY mosquitto.conf /mosquitto/config/mosquitto.conf

# For production, rebuild using mosquitto-secure.conf instead:
# COPY mosquitto-secure.conf /mosquitto/config/mosquitto.conf

# Set proper permissions
RUN chown -R mosquitto:mosquitto /mosquitto

# Expose ports
# 1883 - MQTT (dev), 8883 - MQTT with TLS (prod)
# 9001 - WebSocket (dev), 9443 - WebSocket with TLS (prod)
EXPOSE 1883 8883 9001 9443

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD mosquitto_pub -h localhost -t '$SYS/broker/uptime' -m 'healthcheck' || exit 1

# Use the mosquitto user
USER mosquitto

# Start mosquitto
CMD ["/usr/sbin/mosquitto", "-c", "/mosquitto/config/mosquitto.conf"]
