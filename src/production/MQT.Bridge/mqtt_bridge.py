#!/usr/bin/env python3
"""
MQTT Bridge - Forwards messages from external broker to local Docker broker
This solves the network connectivity issue between Docker and external broker
"""

import json
import os
import time
import sys
from dataclasses import dataclass
from datetime import datetime

try:
    import paho.mqtt.client as mqtt
except ImportError:
    print("❌ paho-mqtt not installed. Install with: pip install paho-mqtt")
    sys.exit(1)


@dataclass
class BridgeConfig:
    """Configuration for the MQTT Bridge (injectable for tests)"""
    external_broker: str
    external_port: int
    local_broker: str
    local_port: int
    topic_filter: str
    external_user: str = ""
    external_pass: str = ""
    local_user: str = ""
    local_pass: str = ""

    @classmethod
    def from_env(cls) -> "BridgeConfig":
        return cls(
            external_broker=os.getenv("EXTERNAL_BROKER_HOST", "172.24.131.97"),
            external_port=int(os.getenv("EXTERNAL_BROKER_PORT", "1883")),
            local_broker=os.getenv("LOCAL_BROKER_HOST", "mosquitto"),
            local_port=int(os.getenv("LOCAL_BROKER_PORT", "1883")),
            topic_filter=os.getenv("TOPIC_FILTER", "sensors/#"),
            external_user=os.getenv("EXTERNAL_BROKER_USER", ""),
            external_pass=os.getenv("EXTERNAL_BROKER_PASS", ""),
            local_user=os.getenv("LOCAL_BROKER_USER", ""),
            local_pass=os.getenv("LOCAL_BROKER_PASS", ""),
        )


class MQTTBridge:
    def __init__(self, config: BridgeConfig = None, external_client=None, local_client=None):
        """
        Initialize the MQTT Bridge.
        Args:
            config: Bridge configuration (default: from env)
            external_client: Optional pre-created external MQTT client (for tests)
            local_client: Optional pre-created local MQTT client (for tests)
        """
        self.config = config or BridgeConfig.from_env()
        self.external_client = external_client
        self.local_client = local_client
        self.connected_to_external = False
        self.connected_to_local = False
        self.running = False
        self._clients_provided = external_client is not None and local_client is not None
        
    def on_external_connect(self, client, userdata, flags, rc, properties=None):
        if rc == 0:
            print(f"✅ Connected to external broker {self.config.external_broker}:{self.config.external_port}")
            self.connected_to_external = True
            # Subscribe to all sensor topics
            client.subscribe(self.config.topic_filter, qos=1)
            print(f"📡 Subscribed to {self.config.topic_filter}")
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
        if self.connected_to_local and self.local_client:
            try:
                # Forward the message to local broker
                result = self.local_client.publish(msg.topic, msg.payload, qos=msg.qos, retain=msg.retain)
                if result.rc == mqtt.MQTT_ERR_SUCCESS:
                    print(f"📤 Forwarded: {msg.topic} -> {self.config.local_broker}")
                else:
                    print(f"❌ Failed to forward message: {result.rc}")
            except Exception as e:
                print(f"❌ Error forwarding message: {e}")
        else:
            print("⚠️  Local broker not connected, dropping message")
    
    def on_local_connect(self, client, userdata, flags, rc, properties=None):
        if rc == 0:
            print(f"✅ Connected to local broker {self.config.local_broker}:{self.config.local_port}")
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
        """Create and configure MQTT clients (skipped if clients provided in __init__)"""
        if self._clients_provided:
            return
        cfg = self.config
        # Create external broker client with new API
        self.external_client = mqtt.Client(mqtt.CallbackAPIVersion.VERSION2,
                                         client_id="mqtt_bridge_external",
                                         clean_session=False)
        self.external_client.on_connect = self.on_external_connect
        self.external_client.on_disconnect = self.on_external_disconnect
        self.external_client.on_message = self.on_external_message

        # Add authentication for external broker if provided
        if cfg.external_user and cfg.external_pass:
            self.external_client.username_pw_set(cfg.external_user, cfg.external_pass)

        # Create local broker client with new API
        self.local_client = mqtt.Client(mqtt.CallbackAPIVersion.VERSION2,
                                      client_id="mqtt_bridge_local",
                                      clean_session=False)
        self.local_client.on_connect = self.on_local_connect
        self.local_client.on_disconnect = self.on_local_disconnect

        # Add authentication for local broker if provided
        if cfg.local_user and cfg.local_pass:
            self.local_client.username_pw_set(cfg.local_user, cfg.local_pass)

        # Enable auto-reconnect
        self.external_client.enable_logger()
        self.local_client.enable_logger()
    
    def connect_clients(self):
        """Connect both clients with retry logic (skipped if clients provided in __init__)"""
        if self._clients_provided:
            return True
        cfg = self.config
        # Connect to external broker
        try:
            print(f"Connecting to external broker {cfg.external_broker}:{cfg.external_port}...")
            self.external_client.connect(cfg.external_broker, cfg.external_port, keepalive=60)
            self.external_client.loop_start()
        except Exception as e:
            print(f"❌ External broker connection error: {e}")
            return False

        # Connect to local broker
        try:
            print(f"Connecting to local broker {cfg.local_broker}:{cfg.local_port}...")
            self.local_client.connect(cfg.local_broker, cfg.local_port, keepalive=60)
            self.local_client.loop_start()
        except Exception as e:
            print(f"❌ Local broker connection error: {e}")
            return False

        return True
    
    def start(self):
        """Start the MQTT bridge"""
        cfg = self.config
        print("=" * 60)
        print("MQTT Bridge - Forwarding External to Local Broker")
        print("=" * 60)
        print(f"📡 External: {cfg.external_broker}:{cfg.external_port}")
        print(f"🏠 Local: {cfg.local_broker}:{cfg.local_port}")
        print(f"🔍 Filter: {cfg.topic_filter}")
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
    bridge = MQTTBridge(config=BridgeConfig.from_env())
    bridge.start()

if __name__ == "__main__":
    main()
