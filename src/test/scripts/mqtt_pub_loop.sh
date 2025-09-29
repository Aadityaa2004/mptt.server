#!/usr/bin/env bash
set -euo pipefail

# Publish N readings with interval seconds between them
# Env: COUNT, INTERVAL, PI_ID, DEVICE_ID, MQTT_HOST, MQTT_PORT, MQTT_USER, MQTT_PASS

COUNT=${COUNT:-10}
INTERVAL=${INTERVAL:-1}
MQTT_HOST=${MQTT_HOST:-localhost}
MQTT_PORT=${MQTT_PORT:-1883}
MQTT_USER=${MQTT_USER:-}
MQTT_PASS=${MQTT_PASS:-}
PI_ID=${PI_ID:-pi_demo_001}
DEVICE_ID=${DEVICE_ID:-101}
METRIC=${METRIC:-reading}

for ((i=1; i<=COUNT; i++)); do
  TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
  VALUE=$(awk -v seed=$RANDOM 'BEGIN{srand(seed); printf "%.2f", 20+10*rand()}')
  PAYLOAD=$(jq -n --arg ts "$TS" --argjson v "$VALUE" '{value: $v, ts: $ts}')
  TOPIC=sensors/${PI_ID}/${DEVICE_ID}/${METRIC}
  echo "[$i/$COUNT] topic=${TOPIC} payload=${PAYLOAD}"
  if [[ -n "$MQTT_USER" ]]; then
    mosquitto_pub -h "$MQTT_HOST" -p "$MQTT_PORT" -u "$MQTT_USER" -P "$MQTT_PASS" -t "$TOPIC" -m "$PAYLOAD" -q 1
  else
    mosquitto_pub -h "$MQTT_HOST" -p "$MQTT_PORT" -t "$TOPIC" -m "$PAYLOAD" -q 1
  fi
  sleep "$INTERVAL"
done

echo "Loop publish complete."


