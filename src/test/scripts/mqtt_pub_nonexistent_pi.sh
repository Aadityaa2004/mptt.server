#!/usr/bin/env bash
set -euo pipefail

# Publish using a bogus PI_ID to verify server rejects when Pi doesn't exist

MQTT_HOST=${MQTT_HOST:-localhost}
MQTT_PORT=${MQTT_PORT:-1883}
MQTT_USER=${MQTT_USER:-}
MQTT_PASS=${MQTT_PASS:-}

PI_ID=${PI_ID:-pi_does_not_exist}
DEVICE_ID=${DEVICE_ID:-999}
METRIC=${METRIC:-reading}
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
PAYLOAD=${PAYLOAD:-$(jq -n --arg ts "$TS" '{value: 1, ts: $ts}')}
TOPIC=sensors/${PI_ID}/${DEVICE_ID}/${METRIC}

echo "Publishing with nonexistent PI_ID: ${PI_ID}"
if [[ -n "$MQTT_USER" ]]; then
  mosquitto_pub -h "$MQTT_HOST" -p "$MQTT_PORT" -u "$MQTT_USER" -P "$MQTT_PASS" -t "$TOPIC" -m "$PAYLOAD" -q 1 || true
else
  mosquitto_pub -h "$MQTT_HOST" -p "$MQTT_PORT" -t "$TOPIC" -m "$PAYLOAD" -q 1 || true
fi

echo "Published. Server should skip this reading; verify logs/DB."


