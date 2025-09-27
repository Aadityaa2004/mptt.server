#!/bin/bash

# MQTT Message Testing Script
# This script publishes test messages to the MQTT broker

set -e

echo "🧪 Testing MQTT message publishing..."

# Check if mosquitto_pub is available
if ! command -v mosquitto_pub &> /dev/null; then
    echo "❌ mosquitto_pub is not installed. Please install mosquitto-clients:"
    echo "   macOS: brew install mosquitto"
    echo "   Ubuntu: sudo apt-get install mosquitto-clients"
    exit 1
fi

# Check if MQTT broker is running
if ! nc -z localhost 1883; then
    echo "❌ MQTT broker is not running on localhost:1883"
    echo "   Please start the services with: docker-compose up -d"
    exit 1
fi

echo "✅ MQTT broker is accessible"

# Function to publish a message
publish_message() {
    local topic=$1
    local message=$2
    local description=$3
    
    echo "📤 Publishing to $topic: $description"
    if mosquitto_pub -h localhost -t "$topic" -m "$message"; then
        echo "✅ Successfully published to $topic"
    else
        echo "❌ Failed to publish to $topic"
        return 1
    fi
    echo ""
}

# Test messages
echo "🚀 Publishing test messages..."

# Temperature readings
publish_message "sensors/temperature" '{"sensor_id":"temp_001","value":25.5,"timestamp":"2024-01-01T12:00:00Z","unit":"celsius"}' "Temperature reading"
publish_message "sensors/temperature" '{"sensor_id":"temp_002","value":23.8,"timestamp":"2024-01-01T12:01:00Z","unit":"celsius"}' "Temperature reading 2"

# Humidity readings
publish_message "sensors/humidity" '{"sensor_id":"hum_001","value":60.2,"timestamp":"2024-01-01T12:00:00Z","unit":"percent"}' "Humidity reading"
publish_message "sensors/humidity" '{"sensor_id":"hum_002","value":58.7,"timestamp":"2024-01-01T12:01:00Z","unit":"percent"}' "Humidity reading 2"

# Pressure readings
publish_message "sensors/pressure" '{"sensor_id":"press_001","value":1013.25,"timestamp":"2024-01-01T12:00:00Z","unit":"hPa"}' "Pressure reading"
publish_message "sensors/pressure" '{"sensor_id":"press_002","value":1012.80,"timestamp":"2024-01-01T12:01:00Z","unit":"hPa"}' "Pressure reading 2"

# Light readings
publish_message "sensors/light" '{"sensor_id":"light_001","value":450.0,"timestamp":"2024-01-01T12:00:00Z","unit":"lux"}' "Light reading"

# Motion detection
publish_message "sensors/motion" '{"sensor_id":"motion_001","value":true,"timestamp":"2024-01-01T12:00:00Z","unit":"boolean"}' "Motion detection"

# Air quality
publish_message "sensors/air_quality" '{"sensor_id":"aq_001","value":45,"timestamp":"2024-01-01T12:00:00Z","unit":"aqi"}' "Air quality reading"

echo "✅ All test messages published successfully!"
echo ""
echo "🔍 To verify messages were processed:"
echo "   1. Check ingestor logs: docker-compose logs ingestor"
echo "   2. Check MongoDB: docker exec -it mongodb mongosh -u admin -p password"
echo "   3. In MongoDB shell: use iot; db.readings.find().pretty()"
echo ""
echo "📊 To monitor MQTT messages in real-time:"
echo "   mosquitto_sub -h localhost -t 'sensors/#'"
