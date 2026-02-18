#!/usr/bin/env python3
"""
MQTT Ingest -> Parse Maple Payload -> Print JSON to stdout
f = temperature
d = distance
"""

import json
import time
import paho.mqtt.client as mqtt
from datetime import datetime

# ----------------------------
# ESP32 CONFIGURATION
# ----------------------------

LOCAL_HOST = "localhost"
LOCAL_PORT = 1883
LOCAL_SUB_TOPIC = "maple/#"


# ---------------------------
# PI CONFIGURATION
# ---------------------------
DEVICE_ID = "PI_001"


# ----------------------------
# PARSER FOR MAPLE PAYLOAD
# ----------------------------

def parse_maple_payload(raw: str):
    """
    Parse payload of form:
    maple/node-MAC id=MAC,f=22.5,d=144
    """
    raw = raw.strip()

    # Split "maple/node-XXXX" from the field section
    try:
        prefix, rest = raw.split(" ", 1)
    except ValueError:
        return {"raw": raw}

    # Extract MAC from prefix
    try:
        node_mac = prefix.split("")[1]
    except:
        node_mac = None

    # Parse id=...,f=...,d=...
    fields = {}
    for pair in rest.split(","):
        if "=" in pair:
            key, value = pair.split("=", 1)
            fields[key.strip()] = value.strip()

    # Convert f → temperature, d → distance
    temp = None
    dist = None

    if "f" in fields:
        try:
            temp = float(fields["f"])
        except:
            temp = fields["f"]

    if "d" in fields:
        try:
            dist = float(fields["d"])
        except:
            dist = fields["d"]

    return {
        "node": node_mac,
        "id": fields.get("id"),
        "pi_id" : DEVICE_ID,
"temperature": temp,
        "distance": dist,
        "ingest_timestamp": datetime.utcnow().isoformat()
    }

# ----------------------------
# LOCAL MQTT INGEST CLIENT
# ----------------------------

ingest = mqtt.Client(client_id="maple-ingest", protocol=mqtt.MQTTv311)

def on_connect(client, userdata, flags, rc):
    print(f"[INGEST] Connected to local MQTT (rc={rc})")
    ingest.subscribe(LOCAL_SUB_TOPIC)
    print(f"[INGEST] Subscribed to: {LOCAL_SUB_TOPIC}")

def on_message(client, userdata, msg):
    raw = msg.payload.decode(errors="ignore")
    print(f"\n[RAW] {msg.topic}: {raw}")

    parsed = parse_maple_payload(raw)
    print("[PARSED]", json.dumps(parsed))

def on_disconnect(client, userdata, rc):
    print(f"[INGEST] Disconnected (rc={rc}), retrying...")
    time.sleep(2)
    try:
        ingest.reconnect()
    except:
        pass

# Register callbacks
ingest.on_connect = on_connect
ingest.on_message = on_message
ingest.on_disconnect = on_disconnect

# Start ingesting
print(f"[START] Connecting to {LOCAL_HOST}:{LOCAL_PORT}...")
ingest.connect(LOCAL_HOST, LOCAL_PORT, keepalive=60)
ingest.loop_forever()