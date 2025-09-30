#!/usr/bin/env python3
"""
MQTT Bridge - Forwards messages from external broker to local Docker broker
This solves the network connectivity issue between Docker and external broker
"""

import json
import time
import sys
from datetime import datetime

try:
    import paho.mqtt.client as mqtt
except ImportError:
    print("❌ paho-mqtt not installed. Install with: pip install paho-mqtt")
    sys.exit(1)

# Configuration
EXTERNAL_BROKER = "172.24.131.97"  # The broker your RPi can reach
EXTERNAL_PORT = 1883
LOCAL_BROKER = "localhost"  # Your Docker broker
LOCAL_PORT = 1883
TOPIC_FILTER = "sensors/#"

class MQTTBridge:
    def __init__(self):
        self.external_client = None
        self.local_client = None
        self.connected_to_external = False
        self.connected_to_local = False
        self.running = False
        
    def on_external_connect(self, client, userdata, flags, rc, properties=None):
        if rc == 0:
            print(f"✅ Connected to external broker {EXTERNAL_BROKER}:{EXTERNAL_PORT}")
            self.connected_to_external = True
            # Subscribe to all sensor topics
            client.subscribe(TOPIC_FILTER, qos=1)
            print(f"📡 Subscribed to {TOPIC_FILTER}")
        else:
            print(f"❌ Failed to connect to external broker, return code {rc}")
            self.connected_to_external = False
    
    def on_external_disconnect(self, client, userdata, rc, properties=None):
        self.connected_to_external = False
        if rc != 0:
            print(f"⚠️  External broker disconnected, return code: {rc}")
            if self.running:
                print("🔄 Will attempt to reconnect...")
        else:
            print("🔌 Disconnected from external broker")
    
    def on_external_message(self, client, userdata, msg):
        """Forward message from external broker to local broker"""
        if self.connected_to_local:
            try:
                # Forward the message to local broker
                result = self.local_client.publish(msg.topic, msg.payload, qos=msg.qos, retain=msg.retain)
                if result.rc == mqtt.MQTT_ERR_SUCCESS:
                    print(f"📤 Forwarded: {msg.topic} -> {LOCAL_BROKER}")
                else:
                    print(f"❌ Failed to forward message: {result.rc}")
            except Exception as e:
                print(f"❌ Error forwarding message: {e}")
        else:
            print("⚠️  Local broker not connected, dropping message")
    
    def on_local_connect(self, client, userdata, flags, rc, properties=None):
        if rc == 0:
            print(f"✅ Connected to local broker {LOCAL_BROKER}:{LOCAL_PORT}")
            self.connected_to_local = True
        else:
            print(f"❌ Failed to connect to local broker, return code {rc}")
            self.connected_to_local = False
    
    def on_local_disconnect(self, client, userdata, rc, properties=None):
        self.connected_to_local = False
        if rc != 0:
            print(f"⚠️  Local broker disconnected, return code: {rc}")
            if self.running:
                print("🔄 Will attempt to reconnect...")
        else:
            print("🔌 Disconnected from local broker")
    
    def create_clients(self):
        """Create and configure MQTT clients"""
        # Create external broker client with new API
        self.external_client = mqtt.Client(mqtt.CallbackAPIVersion.VERSION2, 
                                         client_id="mqtt_bridge_external", 
                                         clean_session=False)
        self.external_client.on_connect = self.on_external_connect
        self.external_client.on_disconnect = self.on_external_disconnect
        self.external_client.on_message = self.on_external_message
        
        # Create local broker client with new API
        self.local_client = mqtt.Client(mqtt.CallbackAPIVersion.VERSION2, 
                                      client_id="mqtt_bridge_local", 
                                      clean_session=False)
        self.local_client.on_connect = self.on_local_connect
        self.local_client.on_disconnect = self.on_local_disconnect
        
        # Enable auto-reconnect
        self.external_client.enable_logger()
        self.local_client.enable_logger()
    
    def connect_clients(self):
        """Connect both clients with retry logic"""
        # Connect to external broker
        try:
            print(f"Connecting to external broker {EXTERNAL_BROKER}:{EXTERNAL_PORT}...")
            self.external_client.connect(EXTERNAL_BROKER, EXTERNAL_PORT, keepalive=60)
            self.external_client.loop_start()
        except Exception as e:
            print(f"❌ External broker connection error: {e}")
            return False
        
        # Connect to local broker
        try:
            print(f"Connecting to local broker {LOCAL_BROKER}:{LOCAL_PORT}...")
            self.local_client.connect(LOCAL_BROKER, LOCAL_PORT, keepalive=60)
            self.local_client.loop_start()
        except Exception as e:
            print(f"❌ Local broker connection error: {e}")
            return False
        
        return True
    
    def start(self):
        """Start the MQTT bridge"""
        print("=" * 60)
        print("MQTT Bridge - Forwarding External to Local Broker")
        print("=" * 60)
        print(f"📡 External: {EXTERNAL_BROKER}:{EXTERNAL_PORT}")
        print(f"🏠 Local: {LOCAL_BROKER}:{LOCAL_PORT}")
        print(f"🔍 Filter: {TOPIC_FILTER}")
        print("Press Ctrl+C to stop...")
        print()
        
        self.running = True
        
        try:
            # Create clients
            self.create_clients()
            
            # Connect clients
            if not self.connect_clients():
                print("❌ Failed to connect to brokers")
                return False
            
            # Wait for connections
            time.sleep(3)
            
            if not self.connected_to_external:
                print("❌ Failed to connect to external broker")
                return False
            
            if not self.connected_to_local:
                print("❌ Failed to connect to local broker")
                return False
            
            print("✅ Bridge is running! Forwarding messages...")
            
            # Keep running with connection monitoring
            while self.running:
                # Check connection status and reconnect if needed
                if not self.connected_to_external:
                    print("🔄 Reconnecting to external broker...")
                    try:
                        self.external_client.reconnect()
                    except Exception as e:
                        print(f"❌ External reconnection failed: {e}")
                
                if not self.connected_to_local:
                    print("🔄 Reconnecting to local broker...")
                    try:
                        self.local_client.reconnect()
                    except Exception as e:
                        print(f"❌ Local reconnection failed: {e}")
                
                time.sleep(5)  # Check every 5 seconds
                
        except KeyboardInterrupt:
            print("\n🛑 Stopping bridge...")
        except Exception as e:
            print(f"❌ Error: {e}")
        finally:
            self.running = False
            self.stop()
    
    def stop(self):
        """Stop the MQTT bridge"""
        if self.external_client:
            self.external_client.loop_stop()
            self.external_client.disconnect()
        if self.local_client:
            self.local_client.loop_stop()
            self.local_client.disconnect()
        print("✅ Bridge stopped")

def main():
    bridge = MQTTBridge()
    bridge.start()

if __name__ == "__main__":
    main()
