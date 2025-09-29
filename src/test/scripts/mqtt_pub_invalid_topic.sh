#!/usr/bin/env bash
set -euo pipefail

# Publish to an invalid topic format to test ingestor validation

MQTT_HOST=${MQTT_HOST:-localhost}
MQTT_PORT=${MQTT_PORT:-1883}
MQTT_USER=${MQTT_USER:-}
MQTT_PASS=${MQTT_PASS:-}

TOPIC=${MQTT_TOPIC:-invalid/topic/without/enough/parts}
PAYLOAD=${PAYLOAD:-'{"test":true}'}

echo "Publishing invalid topic to trigger parser guard: ${TOPIC}"
if [[ -n "$MQTT_USER" ]]; then
  mosquitto_pub -h "$MQTT_HOST" -p "$MQTT_PORT" -u "$MQTT_USER" -P "$MQTT_PASS" -t "$TOPIC" -m "$PAYLOAD" -q 1 || true
else
  mosquitto_pub -h "$MQTT_HOST" -p "$MQTT_PORT" -t "$TOPIC" -m "$PAYLOAD" -q 1 || true
fi

echo "Done. Check server logs for validation message."


