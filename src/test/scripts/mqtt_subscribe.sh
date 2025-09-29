#!/usr/bin/env bash
set -euo pipefail

# Subscribe to the sensors topic tree

MQTT_HOST=${MQTT_HOST:-localhost}
MQTT_PORT=${MQTT_PORT:-1883}
MQTT_USER=${MQTT_USER:-}
MQTT_PASS=${MQTT_PASS:-}
TOPIC=${MQTT_TOPIC:-sensors/#}

echo "Subscribing to ${TOPIC} at ${MQTT_HOST}:${MQTT_PORT}"

if [[ -n "$MQTT_USER" ]]; then
  mosquitto_sub -h "$MQTT_HOST" -p "$MQTT_PORT" -u "$MQTT_USER" -P "$MQTT_PASS" -t "$TOPIC" -v | cat
else
  mosquitto_sub -h "$MQTT_HOST" -p "$MQTT_PORT" -t "$TOPIC" -v | cat
fi


