#!/usr/bin/env python3
"""
Raspberry Pi MQTT Translator
- Subscribes to Maple topics (maple/#)
- Parses Maple payloads and republishes to sensors/{PI_ID}/{DEVICE_ID}/reading
- Robust broker discovery (localhost -> common hostnames -> optional LAN scan)
- Clean shutdown on Ctrl+C
"""

import os
import sys
import time
import json
import socket
from datetime import datetime

try:
    import paho.mqtt.client as mqtt
except ImportError:
    print("❌ paho-mqtt not installed. Install with: pip install paho-mqtt")
    sys.exit(1)

# Configuration
PI_ID = "pi_test"
MAPLE_SUB_TOPIC = "maple/#"

def parse_maple_payload(raw: str):
    """
    Parse payload of form:
    maple/node-MAC: id=MAC,f=22.5,d=144
    or
    maple/node-MAC id=MAC,f=22.5,d=144
    """
    raw = raw.strip()

    # Split "maple/node-XXXX" from the field section (handle both space and colon)
    rest = None
    prefix = None
    if ": " in raw:
        try:
            prefix, rest = raw.split(": ", 1)
        except ValueError:
            return None
    elif " " in raw:
        try:
            prefix, rest = raw.split(" ", 1)
        except ValueError:
            return None
    else:
        # If no separator, try to parse the whole thing as fields
        rest = raw

    # Extract MAC from prefix (maple/node-MAC) if we have a prefix
    node_mac = None
    if prefix:
        try:
            if "-" in prefix:
                node_mac = prefix.split("-")[1]
        except:
            pass

    # Parse id=...,f=...,d=...
    fields = {}
    if rest:
        for pair in rest.split(","):
            if "=" in pair:
                key, value = pair.split("=", 1)
                fields[key.strip()] = value.strip()

    # Extract MAC from id field (this is the DEVICE_ID)
    device_id = fields.get("id")
    if not device_id:
        device_id = node_mac

    # Convert f → temperature, d → distance/level
    temp = None
    dist = None

    if "f" in fields:
        try:
            temp = float(fields["f"])
        except:
            pass

    if "d" in fields:
        try:
            dist = float(fields["d"])
        except:
            pass

    return {
        "device_id": device_id,
        "temperature": temp,
        "distance": dist,
    }

class MqttTranslator:
    def __init__(self, host="localhost", port=1883):
        self.host = host
        self.port = port
        self.connected = False
        self.running = False
        
        # Create client with MQTT v3.1.1 (compatible with most brokers)
        self.client = mqtt.Client(
            client_id=f"translator-{PI_ID}",
            protocol=mqtt.MQTTv311,
            transport="tcp",
            userdata=None,
        )
        
        # Set callbacks
        self.client.on_connect = self.on_connect
        self.client.on_message = self.on_message
        self.client.on_publish = self.on_publish
        self.client.on_disconnect = self.on_disconnect
        
        # Enable logging
        self.client.enable_logger()

    def on_connect(self, client, userdata, flags, rc, *args, **kwargs):
        """
        Handle connection callback for MQTT v3.1.1
        Signature: (client, userdata, flags, rc)
        """
        if rc == 0:
            print(f"✅ Connected to MQTT broker {self.host}:{self.port}")
            self.connected = True
            # Subscribe to Maple topics
            client.subscribe(MAPLE_SUB_TOPIC)
            print(f"📥 Subscribed to: {MAPLE_SUB_TOPIC}")
        else:
            print(f"❌ Failed to connect, return code: {rc}")
            self.connected = False
    
    def on_message(self, client, userdata, msg):
        """
        Handle incoming Maple messages
        """
        try:
            raw = msg.payload.decode(errors="ignore")
            print(f"\n[RAW] {msg.topic}: {raw}")
            
            # Parse Maple payload
            parsed = parse_maple_payload(raw)
            if not parsed or not parsed.get("device_id"):
                print("⚠️  Could not parse device_id from payload, skipping")
                return
            
            device_id = parsed["device_id"]
            temp = parsed.get("temperature")
            dist = parsed.get("distance")
            
            # Publish to sensors topic
            self.publish_sensor_reading(device_id, temp, dist)
            
        except Exception as e:
            print(f"❌ Error processing message: {e}")
    
    def on_publish(self, client, userdata, mid, *args):
        """
        Handle publish callback for MQTT v3.1.1
        Signature: (client, userdata, mid)
        """
        # For MQTT v3.1.1, we don't get reasonCode, just acknowledge
        print(f"🟢 Message published (mid={mid})")
    
    def on_disconnect(self, client, userdata, rc, *args):
        """
        Handle disconnect callback for MQTT v3.1.1
        Signature: (client, userdata, rc)
        """
        self.connected = False
        if rc != 0:
            print(f"⚠️  Disconnected from broker, return code: {rc}")
        else:
            print("🔌 Disconnected from broker")
    
    def connect(self):
        """Connect to MQTT broker"""
        try:
            print(f"Connecting to {self.host}:{self.port}...")
            self.client.connect(self.host, self.port, keepalive=60)
            self.client.loop_start()  # Start background thread
            
            # Wait for connection
            timeout = 10
            start_time = time.time()
            while not self.connected and (time.time() - start_time) < timeout:
                time.sleep(0.1)
            
            if not self.connected:
                print("❌ Connection timeout")
                return False
                
            return True
        except Exception as e:
            print(f"❌ Connection error: {e}")
            return False

    def publish_sensor_reading(self, device_id, temperature, distance):
        """Publish sensor reading in mqtt_envelope format"""
        if not self.connected:
            print("⚠️  Not connected, skipping publish")
            return False
        
        if not device_id:
            print("⚠️  No device_id provided, skipping publish")
            return False

        timestamp = datetime.now().isoformat()
        topic = f"sensors/{PI_ID}/{device_id}/reading"

        # Build sensors dict
        sensors = {}
        if temperature is not None:
            # Temperature from Maple is already in Fahrenheit
            sensors["temperature"] = {
                "value": round(temperature, 2),
                "unit": "fahrenheit",
            }
        
        if distance is not None:
            sensors["level"] = {
                "value": round(distance, 2),
                "unit": "centimeter",
            }

        # Create mqtt_envelope payload
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
        
        try:
            # Publish with QoS 1 for reliability
            result = self.client.publish(topic, json.dumps(envelope), qos=1)
            
            if result.rc == mqtt.MQTT_ERR_SUCCESS:
                sensor_str = []
                if temperature is not None:
                    sensor_str.append(f"temp={temperature:.2f}°F")
                if distance is not None:
                    sensor_str.append(f"level={distance:.2f}cm")
                
                print(
                    f"[{datetime.now().strftime('%H:%M:%S')}] "
                    f"📤 Published to {topic}: {', '.join(sensor_str)}"
                )
                return True
            else:
                print(f"❌ Publish failed: {result.rc}")
                return False
                
        except Exception as e:
            print(f"❌ Publish error: {e}")
            return False

    def start(self):
        """Start the translator (subscribe and translate messages)"""
        self.running = True
        print(f"📡 Translator running - listening to {MAPLE_SUB_TOPIC}")
        print("Press Ctrl+C to stop...")
        
        try:
            # Keep running until interrupted
            while self.running:
                time.sleep(1)
                
        except KeyboardInterrupt:
            print("\n🛑 Stopping...")
            self.running = False
        except Exception as e:
            print(f"❌ Error: {e}")
        finally:
            self.stop()
    
    def stop(self):
        """Stop the translator and disconnect"""
        self.running = False
        try:
            if self.client:
                self.client.loop_stop()   # Stop background thread first
                self.client.disconnect()
        except Exception as e:
            print(f"⚠️  Error during shutdown: {e}")
        print("✅ Translator stopped")

def get_local_ip():
    """Get local IP address"""
    try:
        # Connect to a remote address to determine local IP
        with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as s:
            s.connect(("8.8.8.8", 80))
            return s.getsockname()[0]
    except:
        return "127.0.0.1"

def test_broker_connection(host, port, timeout=3):
    """Test if MQTT broker is reachable"""
    try:
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
            sock.settimeout(timeout)
            result = sock.connect_ex((host, port))
            return result == 0
    except:
        return False

def discover_broker():
    """Discover MQTT broker using multiple strategies"""
    print("🔍 Discovering MQTT broker...")
    
    local_ip = get_local_ip()
    print(f"Local IP: {local_ip}")
    
    # Common broker locations to test
    test_hosts = [
        "localhost",
        "127.0.0.1",
        "mosquitto",
        "broker",
        "mqtt",
        local_ip,
        f"{local_ip.split('.')[0]}.{local_ip.split('.')[1]}.{local_ip.split('.')[2]}.1",  # Gateway
    ]
    
    test_ports = [1883, 8883, 9001]
    
    print("Testing common hosts...")
    
    for host in test_hosts:
        for port in test_ports:
            print(f"Testing {host}:{port}...", end=" ")
            if test_broker_connection(host, port):
                print("✅")
                return host, port
            else:
                print("❌")
    
    print("❌ No broker found. Please check your MQTT broker configuration.")
    return None, None

def main():
    """Main function"""
    print("=" * 60)
    print("Raspberry Pi MQTT Translator")
    print("=" * 60)
    
    # Discover broker
    host, port = discover_broker()
    if not host:
        print("❌ Cannot find MQTT broker")
        return 1
    
    print(f"✅ Found MQTT broker at {host}:{port}")
    
    # Create and start translator
    translator = MqttTranslator(host, port)
    
    if not translator.connect():
        print("❌ Failed to connect to broker")
        return 1
    
    # Start translator (will subscribe and process messages)
    translator.start()
    
    return 0

if __name__ == "__main__":
    sys.exit(main())

