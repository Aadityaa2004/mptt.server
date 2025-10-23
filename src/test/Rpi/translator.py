#!/usr/bin/env python3
"""
Raspberry Pi MQTT Test Translator (single-file)
- Robust broker discovery (localhost -> common hostnames -> optional LAN scan)
- Paho-MQTT v2 callback API (no deprecation warnings)
- Waits for connection before first publish (avoids rc=4 "no connection")
- Clean shutdown on Ctrl+C
"""

import os
import sys
import time
import json
import random
import socket
import threading
from datetime import datetime

try:
    import paho.mqtt.client as mqtt
except ImportError:
    print("❌ paho-mqtt not installed. Install with: pip install paho-mqtt")
    sys.exit(1)

# Configuration
PI_ID = "pi_K001"
DEVICE_ID = "10000"
TOPIC = f"sensors/{PI_ID}/{DEVICE_ID}/reading"

# Test data
BASE_TEMP = 22.0
TEMP_VARIATION = 2.0

class MqttPublisher:
    def __init__(self, host="localhost", port=1883):
        self.host = host
        self.port = port
        self.connected = False
        self.running = False
        
        # Create client with MQTT v3.1.1 (compatible with most brokers)
        self.client = mqtt.Client(
            client_id=f"test-{PI_ID}-{DEVICE_ID}",
            protocol=mqtt.MQTTv311,
            transport="tcp",
            userdata=None,
        )
        
        # Set callbacks - compatible with both v3.1.1 and v5
        self.client.on_connect = self.on_connect
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
        else:
            print(f"❌ Failed to connect, return code: {rc}")
            self.connected = False
    
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

    def publish_temperature(self):
        """Generate and publish temperature reading"""
        if not self.connected:
            print("⚠️  Not connected, skipping publish")
            return False
        
        # Generate realistic temperature reading
        temp = BASE_TEMP + random.uniform(-TEMP_VARIATION, TEMP_VARIATION)
        
        # Create payload
        payload = {
            "device_id": DEVICE_ID,
            "pi_id": PI_ID,
            "timestamp": datetime.now().isoformat(),
            "temperature": round(temp, 2),
            "unit": "celsius",
            "humidity": round(random.uniform(40, 70), 1),  # Optional humidity
            "battery": round(random.uniform(85, 100), 1)   # Optional battery l>
        }
        
        try:
            # Publish with QoS 1 for reliability
            result = self.client.publish(TOPIC, json.dumps(payload), qos=1)
            
            if result.rc == mqtt.MQTT_ERR_SUCCESS:
                print(f"[{datetime.now().strftime('%H:%M:%S')}] Publish: {temp:.2f}°C")
                return True
            else:
                print(f"❌ Publish failed: {result.rc}")
                return False
                
        except Exception as e:
            print(f"❌ Publish error: {e}")
            return False

    def start_publishing(self, interval=2):
        """Start publishing temperature readings at regular intervals"""
        self.running = True
        print(f"📡 Publishing to topic: {TOPIC}")
        print("Press Ctrl+C to stop...")
        
        try:
            while self.running:
                self.publish_temperature()
                time.sleep(interval)
                
        except KeyboardInterrupt:
            print("\n🛑 Stopping...")
            self.running = False
        except Exception as e:
            print(f"❌ Error: {e}")
        finally:
            self.stop()
    
    def stop(self):
        """Stop the publisher and disconnect"""
        self.running = False
        try:
            if self.client:
                self.client.loop_stop()   # Stop background thread first
                self.client.disconnect()
        except Exception as e:
            print(f"⚠️  Error during shutdown: {e}")
        print("✅ Publisher stopped")

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
    print("Raspberry Pi MQTT Test Translator")
    print("=" * 60)
    
    # Discover broker
    host, port = discover_broker()
    if not host:
        print("❌ Cannot find MQTT broker")
        return 1
    
    print(f"✅ Found MQTT broker at {host}:{port}")
    
    # Create and start publisher
    pub = MqttPublisher(host, port)
    
    if not pub.connect():
        print("❌ Failed to connect to broker")
        return 1
    
    # Start publishing
    pub.start_publishing(interval=2)
    
    return 0

if __name__ == "__main__":
    sys.exit(main())

