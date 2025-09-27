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
                                               │     SQL         │
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
- SQL (Postgres)
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


### **Step-by-Step Instructions:**

#### **Step 1: Open pgAdmin**
1. Open your web browser
2. Type in the address bar: `http://localhost:5050`
3. Press Enter
4. You'll see a login page

#### **Step 2: Login to pgAdmin**
1. In the **Email** field, type: `admin@example.com`
2. In the **Password** field, type: `admin`
3. Click **"Login"**

#### **Step 3: Create a Database Connection**
After logging in, you'll see a dashboard. Now you need to tell pgAdmin where your database is:

1. **Find the left sidebar** - you'll see a tree-like structure
2. **Right-click on "Servers"** (it might be collapsed, so expand it first)
3. **Select "Create" → "Server..."** from the menu
4. A new window will open called "Create - Server"

#### **Step 4: Configure the Connection**
The "Create - Server" window has tabs at the top. You need to fill out two tabs:

**General Tab (Step 4a):**
1. Click on the **"General"** tab (it's already selected)
2. In the **"Name"** field, type: `IoT Database`
3. Leave everything else as default

**Connection Tab (Step 4b):**
1. Click on the **"Connection"** tab (next to General)
2. Fill in these fields exactly:
   - **Host name/address**: `postgres`
   - **Port**: `5432`
   - **Maintenance database**: `iot`
   - **Username**: `iot_user`
   - **Password**: `iot_password`
3. Click **"Save"** at the bottom

#### **Step 5: Navigate to Your Data**
After saving, you'll see your database in the left sidebar:

1. **Expand the tree** by clicking the arrows:
   - Click arrow next to **"Servers"**
   - Click arrow next to **"IoT Database"**
   - Click arrow next to **"Databases"**
   - Click arrow next to **"iot"**
   - Click arrow next to **"Schemas"**
   - Click arrow next to **"public"**
   - Click arrow next to **"Tables"**

2. **You'll see 3 tables**:
   - `devices` (information about your devices)
   - `pis` (information about Raspberry Pi gateways)
   - `readings` (actual sensor data)

#### **Step 6: View Your Data**
To see the actual data:

1. **Right-click on any table** (like `readings`)
2. **Select "View/Edit Data" → "All Rows"**
3. You'll see your data in a spreadsheet-like format!

---

## 🧪 **PART 3: Test Everything Works**

### **Complete Test Process:**

#### **Test 1: Send Data via MQTT**
1. Go to `http://localhost:4000` (MQTT Explorer)
2. Make sure you're subscribed to `sensors/#`
3. Publish a test message:
   - Topic: `sensors/pi_test/light/lux`
   - Message: `{"value": 450, "unit": "lux", "timestamp": "2024-01-01T12:00:00Z"}`
4. You should see the message appear in your subscription

#### **Test 2: View Data in Database**
1. Go to `http://localhost:5050` (pgAdmin)
2. Navigate to your `readings` table (see Part 2, Step 6)
3. You should see your test message data in the table!

---

## 🔍 **Understanding Your Data Structure**

### **Your Database Has 3 Tables:**

#### **1. `pis` Table** - Raspberry Pi Gateways
- Like a list of all your Raspberry Pi computers
- Each row is one Pi (pi_001, pi_002, etc.)

#### **2. `devices` Table** - Devices on Each Pi
- Like a list of sensors connected to each Pi
- Each row is one sensor (temperature, humidity, pressure, etc.)

#### **3. `readings` Table** - Actual Sensor Data
- Like a log of all sensor measurements
- Each row is one measurement with the actual values

### **Example Data Flow:**
```
Pi pi_001 → Device "temperature" → Reading {"value": 22.5, "unit": "celsius"}
Pi pi_001 → Device "humidity" → Reading {"value": 65, "unit": "percent"}
Pi pi_002 → Device "pressure" → Reading {"value": 1013.25, "unit": "hPa"}
```

---
