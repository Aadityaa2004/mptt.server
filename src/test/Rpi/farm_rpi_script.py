#!/usr/bin/env python3
"""
farm_rpi_script.py

Runs on the FARM RPi.

- Subscribes to local Maple messages:   maple/#
- Connects to main broker via WSS:      wss://orpheus-networks.com/mqtt
- Publishes MQTT envelope messages to:  sensors/pi_test/<device_id>/reading
"""

import json
import sys
import time
from datetime import datetime

try:
    import paho.mqtt.client as mqtt
except ImportError:
    print("❌ paho-mqtt not installed. Install with: pip install paho-mqtt")
    sys.exit(1)

# ---------------------------------------------------------------------
# CONFIGURATION
# ---------------------------------------------------------------------

# Logical PI ID in your system
PI_ID = "pi_test"

# Local broker on the farm RPi (where Maple node publishes)
LOCAL_MQTT_HOST = "localhost"
LOCAL_MQTT_PORT = 1883
LOCAL_SUB_TOPIC = "maple/#"

# Remote/main broker exposed via Cloudflare/NGINX as WebSocket
REMOTE_MQTT_HOST = "orpheus-networks.com"
REMOTE_MQTT_PORT = 443          # WSS over 443
REMOTE_MQTT_PATH = "/mqtt"      # nginx location /mqtt -> mosquitto:9001
REMOTE_USE_TLS = True           # wss://

# Optional auth if you ever protect the external MQTT
REMOTE_MQTT_USER = None
REMOTE_MQTT_PASS = None

# ---------------------------------------------------------------------
# MAPLE PAYLOAD PARSING
# ---------------------------------------------------------------------

def parse_maple_payload(raw: str):
    """
    Parse payload of form:
      maple/node-MAC: id=MAC,f=22.5,d=144
    or:
      maple/node-MAC id=MAC,f=22.5,d=144
    or just:
      id=MAC,f=...,d=...

    Returns dict with:
      {
        "device_id": "<MAC string>",
        "temperature": <float or None>,
        "distance": <float or None>,
      }
    """
    raw = raw.strip()

    prefix = None
    rest = None

    if ": " in raw:
        # "maple/node-XXXX: id=...,f=...,d=..."
        try:
            prefix, rest = raw.split(": ", 1)
        except ValueError:
            rest = raw
    elif " " in raw:
        # "maple/node-XXXX id=...,f=...,d=..."
        try:
            prefix, rest = raw.split(" ", 1)
        except ValueError:
            rest = raw
    else:
        # No prefix, just the fields
        rest = raw

    node_mac = None
    if prefix:
        # prefix like "maple/node-98A3168FD12C"
        try:
            if "-" in prefix:
                node_mac = prefix.split("-")[1]
        except Exception:
            pass

    fields = {}
    if rest:
        for pair in rest.split(","):
            if "=" in pair:
                key, value = pair.split("=", 1)
                fields[key.strip()] = value.strip()

    device_id = fields.get("id") or node_mac

    # f = temperature (F), d = distance/level (cm)
    temp = None
    dist = None

    if "f" in fields:
        try:
            temp = float(fields["f"])
        except Exception:
            pass

    if "d" in fields:
        try:
            dist = float(fields["d"])
        except Exception:
            pass

    if not device_id:
        return None

    return {
        "device_id": device_id,
        "temperature": temp,
        "distance": dist,
    }

# ---------------------------------------------------------------------
# REMOTE (MAIN PI) MQTT CLIENT
# ---------------------------------------------------------------------

class RemotePublisher:
    def __init__(self):
        # MQTT v3.1.1 over WebSockets
        self.client = mqtt.Client(
            client_id=f"farm-{PI_ID}",
            protocol=mqtt.MQTTv311,
            transport="websockets",
        )
        # WebSocket path (/mqtt)
        self.client.ws_set_options(path=REMOTE_MQTT_PATH)

        if REMOTE_USE_TLS:
            # Use default system CA certificates
            self.client.tls_set()

        if REMOTE_MQTT_USER and REMOTE_MQTT_PASS:
            self.client.username_pw_set(REMOTE_MQTT_USER, REMOTE_MQTT_PASS)

        self.client.on_connect = self.on_connect
        self.client.on_disconnect = self.on_disconnect
        self.client.on_publish = self.on_publish

        self.connected = False

    def on_connect(self, client, userdata, flags, rc):
        if rc == 0:
            print(f"✅ Connected to REMOTE MQTT (wss://{REMOTE_MQTT_HOST}{REMOTE_MQTT_PATH})")
            self.connected = True
        else:
            print(f"❌ Failed to connect REMOTE MQTT, rc={rc}")
            self.connected = False

    def on_disconnect(self, client, userdata, rc):
        self.connected = False
        if rc != 0:
            print(f"⚠️  REMOTE MQTT disconnected unexpectedly (rc={rc})")
        else:
            print("🔌 REMOTE MQTT disconnected")

    def on_publish(self, client, userdata, mid):
        # Acknowledge publish
        print(f"🟢 Remote message published (mid={mid})")

    def connect(self):
        print(f"🌐 Connecting to REMOTE MQTT wss://{REMOTE_MQTT_HOST}{REMOTE_MQTT_PATH} ...")
        self.client.connect(REMOTE_MQTT_HOST, REMOTE_MQTT_PORT, keepalive=60)
        self.client.loop_start()

        # Wait briefly for connection
        for _ in range(50):
            if self.connected:
                break
            time.sleep(0.1)

        if not self.connected:
            print("❌ Remote MQTT connection timeout or failure")
            return False
        return True

    def publish_reading(self, device_id: str, temperature, distance):
        if not self.connected:
            print("⚠️  Remote MQTT not connected, skipping publish")
            return

        timestamp = datetime.utcnow().isoformat()
        topic = f"sensors/{PI_ID}/{device_id}/reading"

        sensors = {}
        if temperature is not None:
            sensors["temperature"] = {
                "value": round(temperature, 2),
                "unit": "fahrenheit",
            }
        if distance is not None:
            sensors["level"] = {
                "value": round(distance, 2),
                "unit": "centimeter",
            }

        envelope = {
            "mqtt_envelope": {
                "topic": topic,
                "payload": {
                    "device_id": device_id,
                    "pi_id": PI_ID,
                    "timestamp": timestamp,
                    "sensors": sensors,
                },
                "qos": 1,
                "retain": False,
                "message_id": int(time.time() * 1000),
                "duplicate": False,
            }
        }

        payload_str = json.dumps(envelope)
        print(f"[{datetime.utcnow().strftime('%H:%M:%S')}] 📤 Publishing to REMOTE: {topic} -> {payload_str}")
        result = self.client.publish(topic, payload_str, qos=1)

        if result.rc != mqtt.MQTT_ERR_SUCCESS:
            print(f"❌ Remote publish failed rc={result.rc}")

# ---------------------------------------------------------------------
# LOCAL (FARM) MQTT CLIENT
# ---------------------------------------------------------------------

class FarmTranslator:
    def __init__(self, remote_publisher: RemotePublisher):
        self.remote = remote_publisher

        self.local_client = mqtt.Client(
            client_id=f"farm-local-{PI_ID}",
            protocol=mqtt.MQTTv311,
        )

        self.local_client.on_connect = self.on_local_connect
        self.local_client.on_message = self.on_local_message
        self.local_client.on_disconnect = self.on_local_disconnect

    def on_local_connect(self, client, userdata, flags, rc):
        if rc == 0:
            print(f"✅ Connected to LOCAL MQTT {LOCAL_MQTT_HOST}:{LOCAL_MQTT_PORT}")
            client.subscribe(LOCAL_SUB_TOPIC)
            print(f"📥 Subscribed to local topic: {LOCAL_SUB_TOPIC}")
        else:
            print(f"❌ Failed to connect LOCAL MQTT, rc={rc}")

    def on_local_disconnect(self, client, userdata, rc):
        if rc != 0:
            print(f"⚠️  LOCAL MQTT disconnected unexpectedly (rc={rc})")
        else:
            print("🔌 LOCAL MQTT disconnected")

    def on_local_message(self, client, userdata, msg):
        try:
            raw = msg.payload.decode(errors="ignore")
            print(f"\n[LOCAL RAW] {msg.topic}: {raw}")
            parsed = parse_maple_payload(raw)
            if not parsed:
                print("⚠️  Could not parse Maple payload, skipping")
                return

            device_id = parsed["device_id"]
            temp = parsed["temperature"]
            dist = parsed["distance"]

            print(
                f"[PARSED] device_id={device_id}, "
                f"temp={temp if temp is not None else 'None'} F, "
                f"dist={dist if dist is not None else 'None'} cm"
            )

            self.remote.publish_reading(device_id, temp, dist)

        except Exception as e:
            print(f"❌ Error processing local message: {e}")

    def start(self):
        print("=" * 60)
        print("Farm RPi MQTT Bridge")
        print("=" * 60)
        print(f"Local broker:  {LOCAL_MQTT_HOST}:{LOCAL_MQTT_PORT} (subscribe: {LOCAL_SUB_TOPIC})")
        print(f"Remote broker: wss://{REMOTE_MQTT_HOST}{REMOTE_MQTT_PATH} (publish: sensors/{PI_ID}/<device_id>/reading)")
        print()

        # Connect remote first
        if not self.remote.connect():
            print("❌ Cannot start translator without remote MQTT connection")
            return 1

        # Connect to local broker
        print(f"📡 Connecting to LOCAL MQTT {LOCAL_MQTT_HOST}:{LOCAL_MQTT_PORT} ...")
        self.local_client.connect(LOCAL_MQTT_HOST, LOCAL_MQTT_PORT, keepalive=60)

        # Run local loop (blocks)
        print("📡 Translator running - press Ctrl+C to stop")
        try:
            self.local_client.loop_forever()
        except KeyboardInterrupt:
            print("\n🛑 Stopping...")
        finally:
            try:
                self.local_client.disconnect()
            except Exception:
                pass
            try:
                self.remote.client.loop_stop()
                self.remote.client.disconnect()
            except Exception:
                pass
            print("✅ Stopped cleanly")
        return 0

# ---------------------------------------------------------------------
# MAIN
# ---------------------------------------------------------------------

def main():
    remote = RemotePublisher()
    translator = FarmTranslator(remote)
    return translator.start()

if __name__ == "__main__":
    sys.exit(main())