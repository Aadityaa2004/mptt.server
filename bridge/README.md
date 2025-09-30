# MQTT Bridge

This containerized MQTT bridge forwards messages from an external broker (where your RPi sensors connect) to the local Docker Mosquitto broker.

## How it works

```
RPi Sensors → External Broker (172.24.131.97) → Bridge Container → Local Mosquitto → Go Ingestor
```

## Configuration

The bridge is configured via environment variables:

- `EXTERNAL_BROKER_HOST`: IP/hostname of external broker (default: 172.24.131.97)
- `EXTERNAL_BROKER_PORT`: Port of external broker (default: 1883)
- `LOCAL_BROKER_HOST`: Docker service name for local broker (default: mosquitto)
- `LOCAL_BROKER_PORT`: Port of local broker (default: 1883)
- `TOPIC_FILTER`: MQTT topic filter to forward (default: sensors/#)

## Usage

The bridge is automatically started with `docker-compose up`. No manual intervention needed!

## Where should it run?

**✅ Recommended: In Docker container (current setup)**
- Clean integration with docker-compose
- Easy to manage and scale
- Consistent with your architecture

**❌ Not recommended: On RPi**
- RPi has limited resources
- Harder to manage remotely
- Network connectivity issues

**❌ Not recommended: On server directly**
- Breaks containerization benefits
- Manual process management
- Harder to deploy and maintain
