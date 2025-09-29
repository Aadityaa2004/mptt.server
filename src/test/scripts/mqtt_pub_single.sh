#!/usr/bin/env bash
set -euo pipefail

# Publish a single JSON reading to the broker
# Env overrides: MQTT_HOST, MQTT_PORT, MQTT_USER, MQTT_PASS, MQTT_TOPIC, PI_ID, DEVICE_ID

MQTT_HOST=${MQTT_HOST:-localhost}
MQTT_PORT=${MQTT_PORT:-1883}
MQTT_USER=${MQTT_USER:-}
MQTT_PASS=${MQTT_PASS:-}

PI_ID=${PI_ID:-pi_demo_001}
DEVICE_ID=${DEVICE_ID:-101}
METRIC=${METRIC:-reading}

TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
PAYLOAD=${PAYLOAD:-$(jq -n --arg ts "$TS" '{temp: 22.5, humidity: 58, ts: $ts}')}

TOPIC=${MQTT_TOPIC:-sensors/${PI_ID}/${DEVICE_ID}/${METRIC}}

echo "Publishing to ${MQTT_HOST}:${MQTT_PORT} topic=${TOPIC} payload=${PAYLOAD}"

if [[ -n "$MQTT_USER" ]]; then
  mosquitto_pub -h "$MQTT_HOST" -p "$MQTT_PORT" -u "$MQTT_USER" -P "$MQTT_PASS" -t "$TOPIC" -m "$PAYLOAD" -q 1
else
  mosquitto_pub -h "$MQTT_HOST" -p "$MQTT_PORT" -t "$TOPIC" -m "$PAYLOAD" -q 1
fi

echo "Done."


