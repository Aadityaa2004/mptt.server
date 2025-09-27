# MPT.MQTT_Server

This server is designed for high performance MQTT data ingestion. It is built in GO and capable of handling requests from MQTT Brokers and also responsible for making pushes to the database. This service is specifically designed for IoT applications and supports both development and production environments. 

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   IoT Devices   │───▶│  MQTT Broker    │───▶│  Go Ingestor    │
│   (Sensors)     │    │  (Mosquitto)    │    │   Service       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                        │
                                                        ▼
                                               ┌─────────────────┐
                                               │    MongoDB      │
                                               │   Database      │
                                               └─────────────────┘
```

## Features

- **MQTT Data Ingestion**: Subscribes to MQTT topics and processes sensor data
- **Batch Processing**: Efficiently batches data for optimal database performance
- **Health Monitoring**: HTTP endpoints for service health checks
- **Docker Support**: Complete containerization for easy deployment
- **TLS Support**: Secure MQTT connections with certificate validation
- **Scalable Architecture**: Support for shared consumer groups

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.22+ (for local development)
- MongoDB (local or cloud)
- brew install mosquitto

### Local Development with Docker

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd mpt.mqtt_server
   ```

2. **Create environment file**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start all services**
   ```bash
   docker-compose up -d
   ```

4. **Check service health**
   ```bash
   curl http://localhost:9002/health
   ```

5. **View logs**
   ```bash
   docker-compose logs -f ingestor
   ```

## Data Model

### Reading Document Structure

```json
{
  "_id": "ObjectId",
  "topic": "sensors/device001/temperature",
  "device_id": "device001",
  "payload": {
    "temperature": 23.5,
    "unit": "celsius",
    "timestamp": 1640995200
  },
  "received_at": "2023-12-31T12:00:00Z"
}
```

## API Endpoints

### Health Check
- **GET** `/health` - Service health status
- **Response**: `{"status": "ok"}`

## Docker Services

### MQTT Broker (Mosquitto)
- **Image**: `eclipse-mosquitto:2`
- **Port**: 1883 (dev), 8883 (prod with TLS)
- **Configuration**: Supports both anonymous and authenticated connections

### MongoDB
- **Image**: `mongo:7`
- **Port**: 27017
- **Persistence**: Data persisted in Docker volume

### Ingestor Service
- **Build**: Multi-stage Go build
- **Base Image**: `gcr.io/distroless/base-debian12`
- **Port**: 9002 (HTTP health endpoint)

### Testing

- **make test-mqtt** 

## Troubleshooting

### Common Issues

1. **MQTT Connection Failed**
   - Check broker hostname and port
   - Verify network connectivity
   - Check authentication credentials

2. **MongoDB Connection Failed**
   - Verify MongoDB URI
   - Check database permissions
   - Ensure MongoDB is running

3. **High Memory Usage**
   - Reduce batch size
   - Increase batch window
   - Check for memory leaks
