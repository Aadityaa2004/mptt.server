#!/bin/bash

# Test PostgreSQL and mqtt Setup Script
# This script helps you test the PostgreSQL migration

echo "=== PostgreSQL MQTT Server Test Setup ==="
echo

# Check if PostgreSQL is running locally
echo "1. Checking if PostgreSQL is running locally..."
if pg_isready -h localhost -p 5432 > /dev/null 2>&1; then
    echo "✅ PostgreSQL is running on localhost:5432"
else
    echo "❌ PostgreSQL is not running on localhost:5432"
    echo "   Please start PostgreSQL or use Docker Compose"
    echo
    echo "To start with Docker Compose:"
    echo "   docker-compose up -d postgres"
    echo
    echo "To start locally (if installed):"
    echo "   brew services start postgresql"
    echo
fi

echo

# Test database connection
echo "2. Testing database connection..."
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=iot_user
export POSTGRES_PASSWORD=iot_password
export POSTGRES_DB=iot
export POSTGRES_SSLMODE=disable

# Test connection using psql
if psql -h localhost -p 5432 -U iot_user -d iot -c "SELECT 1;" > /dev/null 2>&1; then
    echo "✅ Database connection successful"
else
    echo "❌ Database connection failed"
    echo "   Please check your PostgreSQL credentials and ensure the database exists"
    echo
    echo "To create the database and user:"
    echo "   psql -h localhost -p 5432 -U postgres -c \"CREATE DATABASE iot;\""
    echo "   psql -h localhost -p 5432 -U postgres -c \"CREATE USER iot_user WITH PASSWORD 'iot_password';\""
    echo "   psql -h localhost -p 5432 -U postgres -c \"GRANT ALL PRIVILEGES ON DATABASE iot TO iot_user;\""
    echo
fi

echo

# Test MQTT broker
echo "3. Testing MQTT broker..."
if nc -z localhost 1883 > /dev/null 2>&1; then
    echo "✅ MQTT broker is running on localhost:1883"
else
    echo "❌ MQTT broker is not running on localhost:1883"
    echo "   Please start the MQTT broker or use Docker Compose"
    echo
    echo "To start with Docker Compose:"
    echo "   docker-compose up -d mosquitto"
    echo
fi

echo

# Test the application
echo "4. Testing the application..."
echo "Building the application..."
if go build -o mqtt-ingestor ./src/production/MQT.Startup/; then
    echo "✅ Application built successfully"
    
    echo "Starting the application (will run for 10 seconds)..."
    timeout 10s ./mqtt-ingestor &
    APP_PID=$!
    
    sleep 5
    
    # Test health endpoint
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo "✅ Health endpoint is responding"
    else
        echo "❌ Health endpoint is not responding"
    fi
    
    # Test detailed health endpoint
    if curl -s http://localhost:8080/health/detailed > /dev/null 2>&1; then
        echo "✅ Detailed health endpoint is responding"
        echo "   Response:"
        curl -s http://localhost:8080/health/detailed | jq . 2>/dev/null || curl -s http://localhost:8080/health/detailed
    else
        echo "❌ Detailed health endpoint is not responding"
    fi
    
    # Kill the application
    kill $APP_PID 2>/dev/null
    wait $APP_PID 2>/dev/null
    
    echo "✅ Application test completed"
else
    echo "❌ Failed to build application"
    echo "   Please check for compilation errors"
fi

echo

# Test MQTT message publishing
echo "5. Testing MQTT message publishing..."
echo "Publishing test messages to MQTT broker..."

# Test message format: sensors/<pi_id>/<device_id>/<metric>
mosquitto_pub -h localhost -p 1883 -t "sensors/pi_001/temperature/humidity" -m '{"value": 65.2, "unit": "percent", "timestamp": "2024-01-01T12:00:00Z"}'
mosquitto_pub -h localhost -p 1883 -t "sensors/pi_001/temperature/temp" -m '{"value": 22.5, "unit": "celsius", "timestamp": "2024-01-01T12:00:00Z"}'
mosquitto_pub -h localhost -p 1883 -t "sensors/pi_002/pressure/barometric" -m '{"value": 1013.25, "unit": "hPa", "timestamp": "2024-01-01T12:00:00Z"}'

echo "✅ Test messages published"
echo

# Database verification
echo "6. Verifying data in PostgreSQL..."
echo "Checking if data was inserted into the database..."

# Check if tables exist
if psql -h localhost -p 5432 -U iot_user -d iot -c "\dt" > /dev/null 2>&1; then
    echo "✅ Tables exist in the database"
    
    # Check pis table
    echo "   Pis table:"
    psql -h localhost -p 5432 -U iot_user -d iot -c "SELECT pi_id, created_at FROM pis ORDER BY created_at DESC LIMIT 5;"
    
    # Check devices table
    echo "   Devices table:"
    psql -h localhost -p 5432 -U iot_user -d iot -c "SELECT pi_id, device_id, created_at FROM devices ORDER BY created_at DESC LIMIT 5;"
    
    # Check readings table
    echo "   Readings table:"
    psql -h localhost -p 5432 -U iot_user -d iot -c "SELECT pi_id, device_id, ts, payload FROM readings ORDER BY ts DESC LIMIT 5;"
    
else
    echo "❌ Tables do not exist in the database"
    echo "   The application should create tables automatically on startup"
fi

echo
echo "=== Test Setup Complete ==="
echo
echo "Next steps:"
echo "1. Start the full stack: docker-compose up -d"
echo "2. Monitor logs: docker-compose logs -f ingestor"
echo "3. Test MQTT messages: Use MQTT Explorer or mosquitto_pub"
echo "4. Check database: psql -h localhost -p 5432 -U iot_user -d iot"
echo
echo "Expected MQTT topic format: sensors/<pi_id>/<device_id>/<metric>"
echo "Example: sensors/pi_001/temperature/humidity"
