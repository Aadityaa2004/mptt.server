#!/usr/bin/env bash
set -euo pipefail

# Subscribe to error topics from the ingestor

MQTT_HOST=${MQTT_HOST:-localhost}
MQTT_PORT=${MQTT_PORT:-1883}
MQTT_USER=${MQTT_USER:-}
MQTT_PASS=${MQTT_PASS:-}

# Subscribe to all error topics
TOPIC=${MQTT_TOPIC:-ingestor/errors/#}

echo "Subscribing to error topics: ${TOPIC} at ${MQTT_HOST}:${MQTT_PORT}"
echo "This will show validation errors when Pis publish invalid data"

if [[ -n "$MQTT_USER" ]]; then
  mosquitto_sub -h "$MQTT_HOST" -p "$MQTT_PORT" -u "$MQTT_USER" -P "$MQTT_PASS" -t "$TOPIC" -v | cat
else
  mosquitto_sub -h "$MQTT_HOST" -p "$MQTT_PORT" -t "$TOPIC" -v | cat
fi



