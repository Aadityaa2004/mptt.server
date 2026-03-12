# MPT.MQTT_Server

This project is licensed under the MIT License—see [LICENSE](LICENSE).

A **true microservice architecture** for high-performance MQTT data ingestion with Role-Based Access Control (RBAC). Built in Go, this system handles IoT sensor data from MQTT brokers and provides a complete REST API for data management. Specifically designed for IoT applications, particularly maple syrup farm monitoring systems, with support for both development and production environments.

## 🏗️ **Microservice Architecture Overview**

This system follows **proper microservice principles** with complete decoupling and service independence. Each service has a single responsibility and communicates via well-defined APIs.

## 🔐 Authentication & Authorization

The system uses a simplified RBAC approach with only two roles:

### **Admin Role**
- Full CRUD access to all resources (users, PIs, devices, readings)
- Can create/update/delete any user
- Can assign/change user roles
- Can create/manage PIs and assign them to users
- Can create/manage devices
- Can view all readings

### **User Role**
- Read-only access to their assigned resources
- Can view PIs assigned to them (`pis.user_id = user.user_id`)
- Can view devices connected to their PIs
- Can view readings from their devices
- Can view/update their own profile

### **Token System**
- **Access Token**: Contains `user_id`, `role`, `token_id`, and standard JWT claims
- **Refresh Token**: Used to generate new access tokens
- **No Permission Tokens**: Simplified to role-based authorization only
- **Token Expiration**: Old tokens remain valid until natural expiration 

## 📊 **Current Services (5 Microservices + Infrastructure)**

### **1. MQTT Ingestor Service** (`mqtt-ingestor`)
- **Port**: 9003
- **Purpose**: Pure data ingestion from MQTT
- **Responsibilities**:
  - Subscribe to MQTT broker (`sensors/#` topics)
  - Receive sensor readings from devices
  - Validate Pi/Device existence via API calls
  - Store readings via API calls
  - **NO direct database access**
  - Publish error messages back to MQTT
  - Health check endpoint

### **2. API Service** (`api-service`) 
- **Port**: 9002
- **Purpose**: Single source of truth for all data operations
- **Responsibilities**:
  - All REST API endpoints (users, pis, devices, readings)
  - JWT authentication & RBAC
  - **ALL database operations** (only service that touches PostgreSQL)
  - Internal API endpoints for ingestor service
  - User management and authentication
  - Health check endpoints

### **3. MQTT Bridge** (`mqtt-bridge`)
- **Purpose**: External broker connectivity
- **Responsibilities**:
  - Forward messages from external MQTT broker to internal broker
  - Bridge between external sensors and internal system
  - Topic filtering and message routing

### **4. MQTT Broker** (`mosquitto`)
- **Ports**: 1883 (MQTT), 8883 (TLS), 9001 (WebSocket), 9443 (WebSocket TLS)
- **Purpose**: Internal MQTT message broker
- **Responsibilities**:
  - Message routing and pub/sub
  - Message persistence
  - TLS/SSL support

### **5. PostgreSQL Database** (`postgres`)
- **Port**: 5432
- **Purpose**: Data persistence
- **Responsibilities**:
  - Store all application data
  - **Only accessed by API Service**

## 🔄 **Complete Data Flow**

```
External Sensors → MQTT Bridge → Internal MQTT Broker → MQTT Ingestor → API Service → PostgreSQL
                                                                        ↑
External Users ← REST API ← API Service ← JWT Authentication ← External Users
```

### **Detailed Flow:**

1. **Sensor Data Ingestion**:
   ```
   Raspberry Pi Sensors → External MQTT Broker → MQTT Bridge → Internal Mosquitto → MQTT Ingestor Service
   ```

2. **Data Processing**:
   ```
   MQTT Ingestor → HTTP API Call → API Service → PostgreSQL
   ```

3. **User Access**:
   ```
   Web/Mobile App → REST API → API Service → PostgreSQL
   ```

## 🔐 **Authentication & Authorization**

### **Two Types of Authentication:**

1. **User Authentication** (API Service):
   - JWT tokens for web/mobile users
   - Username/password login
   - Role-based access control (RBAC)
   - Admin user management

2. **Service-to-Service Authentication** (Internal):
   - Bearer token authentication
   - `INTERNAL_API_SECRET` for ingestor ↔ API communication
   - No user credentials needed

## 🎯 **Key Microservice Principles Achieved**

### **✅ Service Independence**
- Each service has its own responsibility
- Services can be deployed independently
- No shared code except data models

### **✅ Database per Service**
- Only API Service accesses PostgreSQL
- Ingestor Service has no database dependencies
- Clear data ownership

### **✅ API-First Communication**
- Services communicate via HTTP APIs
- No direct database sharing
- Well-defined service contracts

### **✅ Fault Tolerance**
- Circuit breaker pattern in API client
- Retry logic with exponential backoff
- Graceful degradation

### **✅ Independent Scaling**
- Scale ingestor for high MQTT throughput
- Scale API for high HTTP traffic
- Different resource requirements

### **No Separate Database Services**
There is **NO separate database service**. PostgreSQL is just infrastructure, and only the API Service accesses it.

## 📋 **Service Communication Matrix**

| Service | Communicates With | Method | Purpose |
|---------|------------------|--------|---------|
| MQTT Ingestor | API Service | HTTP POST | Validate Pi/Device, Create readings |
| MQTT Ingestor | MQTT Broker | MQTT | Subscribe to topics, publish errors |
| MQTT Bridge | External Broker | MQTT | Forward messages from external sensors |
| MQTT Bridge | Internal Broker | MQTT | Forward messages to internal system |
| API Service | PostgreSQL | SQL | All database operations |
| External Users | API Service | HTTP | User authentication, data access |

## 🎉 **Why This IS True Microservice Architecture**

1. **Single Responsibility**: Each service has one clear purpose
2. **Loose Coupling**: Services communicate via APIs, not shared databases
3. **High Cohesion**: Related functionality is grouped together
4. **Independent Deployment**: Services can be deployed separately
5. **Technology Agnostic**: Services could use different technologies
6. **Fault Isolation**: Failure in one service doesn't crash others
7. **Scalability**: Each service can be scaled independently

The architecture is now properly decoupled and follows microservice best practices! 🚀

## Features

### **🏗️ Microservice Architecture**
- **True Service Independence**: Each service has a single responsibility
- **API-First Communication**: Services communicate via HTTP APIs
- **Database per Service**: Clear data ownership and boundaries
- **Independent Deployment**: Services can be deployed separately
- **Fault Tolerance**: Circuit breaker and retry patterns

### **🔐 Authentication & Security**
- **Simplified RBAC**: Two-role system (admin/user) with clear permissions
- **JWT Authentication**: Secure token-based authentication with refresh tokens
- **Service-to-Service Auth**: Internal API authentication with shared secrets
- **TLS Support**: Secure MQTT connections with certificate validation

### **📡 MQTT & Data Processing**
- **MQTT Data Ingestion**: Subscribes to MQTT topics and processes sensor data
- **Batch Processing**: Efficiently batches data for optimal performance
- **External Broker Bridge**: Connects to external MQTT brokers
- **Error Handling**: Publishes error messages back to MQTT topics

### **🔧 Operations & Monitoring**
- **Health Monitoring**: HTTP endpoints for service health checks
- **Circuit Breaker Status**: Real-time monitoring of service resilience
- **Docker Support**: Complete containerization for easy deployment
- **Scalable Architecture**: Independent scaling of services
- **Maple Syrup Farm Ready**: Designed for IoT monitoring of maple syrup production

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
   # API Service (includes authentication)
   curl http://localhost:9002/health/live
   
   # MQTT Ingestor Service
   curl http://localhost:9003/health
   ```

5. **Test authentication**
   ```bash
   # Login as admin (default credentials)
   curl -X POST http://localhost:9002/api/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username": "admin", "password": "adminpassword123"}'
   ```

6. **View logs**
   ```bash
   docker-compose logs -f api-service
   docker-compose logs -f mqtt-ingestor
   docker-compose logs -f mqtt-bridge
   ```

### Local vs production routing

The codebase supports both environments with the same repo:

| Where you run | Compose file | API / frontend base URL | CORS |
|---------------|--------------|-------------------------|------|
| **Dev machine (local)** | `docker-compose.yml` | `http://localhost:9002` | `http://localhost:3000` |
| **RPi / production** | `docker-compose.rpi.yml` + `.env.production` | `https://orpheus-networks.com` | `https://orpheus-networks.com` |

- **Local (`docker-compose up`)**: Frontend is built with `NEXT_PUBLIC_API_BASE_URL=http://localhost:9002`; no nginx; you open the app at `http://localhost:3000` and it talks to the API at `http://localhost:9002`.
- **Production**: Frontend image is built with orpheus URLs (e.g. via `./push-to-dockerhub.sh`); nginx and Cloudflare Tunnel handle routing and TLS.

The frontend fallback in code is `http://localhost:9002` so that unset env = local. Production images are built with explicit `NEXT_PUBLIC_*` URLs.

### Testing the RPi stack locally (same routing as production)

To run the full RPi stack (including nginx) on your dev machine with **localhost** instead of orpheus-networks.com:

1. **Copy the override and local env example**
   ```bash
   cp docker-compose.rpi.override.example.yml docker-compose.rpi.override.yml
   cp env.rpi.local.example .env.rpi.local
   # Edit .env.rpi.local with your DB, JWT, etc. (or keep .env.production and only override will change CORS/URLs)
   ```

2. **Run with the override** (builds frontend with localhost; nginx uses `default-local.conf`)
   ```bash
   docker compose -f docker-compose.rpi.yml -f docker-compose.rpi.override.yml --env-file .env.rpi.local up --build
   ```
   Or if you already have `.env.production` and only want CORS + frontend URLs overridden:
   ```bash
   docker compose -f docker-compose.rpi.yml -f docker-compose.rpi.override.yml up --build
   ```

3. **Open** `http://localhost` — routing matches production (all traffic through nginx), but with localhost and no Cloudflare.

`docker-compose.rpi.override.yml` and `.env.rpi.local` are gitignored so they stay local.

## Data Model

### **Database Schema (PostgreSQL)**

#### **Users Table**
```sql
CREATE TABLE users (
    user_id     TEXT PRIMARY KEY,
    username    TEXT NOT NULL UNIQUE,
    email       TEXT NOT NULL UNIQUE,
    password    TEXT NOT NULL,
    role        TEXT NOT NULL,  -- 'admin' or 'user'
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

#### **Roles Table**
```sql
CREATE TABLE roles (
    role_id     TEXT PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

#### **PIs Table**
```sql
CREATE TABLE pis (
    pi_id       TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    location    TEXT,
    user_id     TEXT NOT NULL,  -- Foreign key to users
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

#### **Devices Table**
```sql
CREATE TABLE devices (
    device_id   TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    type        TEXT NOT NULL,
    pi_id       TEXT NOT NULL,  -- Foreign key to pis
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

#### **Readings Table**
```sql
CREATE TABLE readings (
    reading_id  TEXT PRIMARY KEY,
    device_id   TEXT NOT NULL,  -- Foreign key to devices
    value       DECIMAL NOT NULL,
    unit        TEXT NOT NULL,
    timestamp   TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### **Authentication Response Structure**

#### **Login Response**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_id": "uuid-string",
  "expires_at": 1761176384,
  "user_id": "user-uuid",
  "username": "admin",
  "email": "admin@example.com",
  "role": "admin"
}
```

#### **User Registration Response**
```json
{
  "id": "user-uuid",
  "username": "newuser",
  "email": "newuser@example.com",
  "role": "user"
}
```

## API Endpoints

> **Adding new endpoints?** See [ROUTING.md](ROUTING.md) for nginx routing rules to avoid 404s in production.

### **API Service** (Port 9002) - Single Service for All Operations

#### **Health & Monitoring**
- **GET** `/health/live` - Service liveness check
- **GET** `/health/ready` - Service readiness check
- **GET** `/metrics` - Service metrics
- **GET** `/stats/summary` - System statistics

#### **Authentication & User Management**
- **POST** `/api/auth/login` - User login
- **POST** `/api/auth/register` - User registration
- **GET** `/api/auth/profile` - Get user profile
- **POST** `/api/auth/refresh` - Refresh access token
- **POST** `/api/auth/logout` - User logout
- **GET** `/api/users` - Get all users (Admin only)
- **GET** `/api/users/{id}` - Get user by ID
- **PUT** `/api/users/{id}` - Update user
- **PUT** `/api/users/{id}/role` - Update user role (Admin only)
- **DELETE** `/api/users/{id}` - Delete user (Admin only)

#### **PI Management**
- **POST** `/api/pis` - Create PI (Admin only)
- **GET** `/api/pis` - Get PIs (Admin: all, User: assigned)
- **GET** `/api/pis/{id}` - Get PI details
- **PUT** `/api/pis/{id}` - Update PI (Admin only)
- **DELETE** `/api/pis/{id}` - Delete PI (Admin only)

#### **Device Management**
- **POST** `/api/pis/{pi_id}/devices` - Create device (Admin only)
- **GET** `/api/pis/{pi_id}/devices` - Get devices (Admin: all, User: from assigned PIs)
- **GET** `/api/pis/{pi_id}/devices/{device_id}` - Get device details
- **PUT** `/api/pis/{pi_id}/devices/{device_id}` - Update device (Admin only)
- **DELETE** `/api/pis/{pi_id}/devices/{device_id}` - Delete device (Admin only)

#### **Reading Management**
- **POST** `/api/readings` - Create reading (Admin only)
- **GET** `/api/readings` - Get readings (Admin: all, User: from assigned devices)
- **GET** `/api/readings/latest?pi_id={id}` - Get latest readings
- **GET** `/api/readings/pis/{pi_id}/devices/{device_id}` - Get device readings

#### **Internal API Endpoints** (Service-to-Service)
- **POST** `/internal/pis/validate` - Validate Pi exists (Ingestor → API)
- **POST** `/internal/devices/validate` - Validate Device exists (Ingestor → API)
- **POST** `/internal/readings` - Create readings (Ingestor → API)

### **MQTT Ingestor Service** (Port 9003) - Health Only
- **GET** `/health` - Service health with circuit breaker status

## Docker Services

### **MQTT Broker (Mosquitto)**
- **Image**: `eclipse-mosquitto:2`
- **Ports**: 1883 (MQTT), 8883 (TLS), 9001 (WebSocket), 9443 (WebSocket TLS)
- **Configuration**: Supports both anonymous and authenticated connections
- **Purpose**: Internal MQTT message broker

### **MQTT Bridge**
- **Build**: Python-based bridge service
- **Base Image**: `python:3.11-alpine`
- **Purpose**: Forward messages from external MQTT broker to internal broker
- **Features**: Topic filtering, message routing, external broker connectivity

### **API Service** (Port 9002)
- **Build**: Multi-stage Go build
- **Base Image**: `gcr.io/distroless/base-debian12`
- **Purpose**: Single source of truth for all data operations
- **Features**: 
  - JWT authentication & RBAC
  - All REST API endpoints
  - Database operations
  - Internal API endpoints for ingestor
  - Health monitoring

### **MQTT Ingestor Service** (Port 9003)
- **Build**: Multi-stage Go build
- **Base Image**: `gcr.io/distroless/base-debian12`
- **Purpose**: Pure MQTT data ingestion
- **Features**:
  - MQTT subscription and processing
  - API client with circuit breaker
  - Batch processing
  - Error publishing to MQTT
  - Health monitoring with circuit breaker status

### **PostgreSQL Database**
- **Image**: `postgres:15`
- **Port**: 5432
- **Purpose**: Data persistence
- **Access**: Only by API Service

## 🔧 **Resilience & Circuit Breaker**

### **Circuit Breaker Pattern**
The MQTT Ingestor Service implements a circuit breaker pattern for resilience:

- **States**: Closed (normal), Open (failing), Half-Open (testing)
- **Failure Threshold**: 5 consecutive failures
- **Reset Timeout**: 30 seconds
- **Retry Logic**: Exponential backoff (1s, 2s, 4s, 8s)

### **Health Check with Circuit Breaker Status**
```bash
curl http://localhost:9003/health
```

**Response Example**:
```json
{
  "status": "healthy",
  "timestamp": "2025-10-23T04:09:48Z",
  "services": {
    "mqtt": "connected",
    "api_service": "connected"
  },
  "circuit_breaker": {
    "state": "closed",
    "failure_count": 0
  }
}
```

### **Fault Tolerance Features**
- **Automatic Retry**: Failed API calls are retried with exponential backoff
- **Circuit Breaking**: Prevents cascading failures when API service is down
- **Graceful Degradation**: Ingestor continues to receive MQTT messages even if API is unavailable
- **Error Publishing**: Failed readings are published to MQTT error topics for device feedback

## 🧪 Testing

## 🎯 **Microservice Architecture Benefits**

### **Development Benefits**
- **Independent Development**: Teams can work on services independently
- **Technology Diversity**: Each service can use different technologies
- **Faster Deployment**: Deploy only the services that changed
- **Easier Testing**: Test services in isolation

### **Operational Benefits**
- **Independent Scaling**: Scale services based on their specific needs
- **Fault Isolation**: Failure in one service doesn't crash others
- **Resource Optimization**: Allocate resources based on service requirements
- **Easier Maintenance**: Update services without affecting others

### **Business Benefits**
- **Faster Time to Market**: Deploy features independently
- **Better Reliability**: Circuit breakers prevent cascading failures
- **Cost Optimization**: Scale only what you need
- **Future-Proof**: Easy to add new services or modify existing ones

## 🧪 **Testing**

### **Service Health Testing**
```bash
# Test API Service
curl http://localhost:9002/health/live

# Test MQTT Ingestor Service (with circuit breaker status)
curl http://localhost:9003/health

# Test MQTT Bridge
docker-compose logs mqtt-bridge
```

### **End-to-End Flow Testing**
```bash
# 1. Send MQTT message
docker exec mqtt-broker mosquitto_pub -h localhost -t "sensors/pi_001/1/temperature" -m '{"temperature": 26.0, "humidity": 65.0, "timestamp": "2025-10-23T04:10:00Z"}'

# 2. Check ingestor logs
docker logs mqtt-ingestor --tail 5

# 3. Test API authentication
curl -X POST http://localhost:9002/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "adminpassword123"}'
```

### **Circuit Breaker Testing**
```bash
# Stop API service to test circuit breaker
docker-compose stop api-service

# Send MQTT messages (should trigger circuit breaker)
docker exec mqtt-broker mosquitto_pub -h localhost -t "sensors/pi_001/1/temperature" -m '{"temperature": 26.0}'

# Check circuit breaker status
curl http://localhost:9003/health

# Restart API service
docker-compose start api-service
```

### **Manual Testing**

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

## 🔍 **API Endpoints**

| Controller | Endpoint | Method | Access | Description |
|------------|----------|--------|--------|-------------|
| **auth_controller.go** | | | | **Authentication endpoints** |
| | `/api/auth/register` | POST | Public | User registration (user role forced) |
| | `/api/auth/login` | POST | Public | User login |
| | `/api/auth/refresh` | POST | Public | Refresh token |
| | `/api/auth/logout` | POST | Public | User logout |
| | `/api/auth/profile` | GET | Authenticated | Get own profile |
| | `/api/auth/profile` | PATCH | Authenticated | Update own profile (username, email, password) |
| | `/api/auth/register/admin` | POST | Admin only | Admin registration |
| **user_controller.go** | | | | **User management** |
| | `/api/users` | GET | Admin only | List all users |
| | `/api/users/:id` | GET | Admin or Owner | View user details |
| | `/api/users/:id` | PUT | Admin only | Update any user |
| | `/api/users/:id` | DELETE | Admin only | Hard delete user |
| | `/api/users/:id/role` | PUT | Admin only | Change user role |
| **pi_controller.go** | | | | **Pi management** |
| | `/pis` | POST | Admin only | Create pi, assign to user |
| | `/pis` | GET | Admin: all PIs<br>User: only their assigned PIs | List PIs |
| | `/pis/:pi_id` | GET | Admin: any PI<br>User: only their assigned PI | Get PI details |
| | `/pis/:pi_id` | PATCH | Admin only | Update pi, reassign user |
| | `/pis/:pi_id` | DELETE | Admin only | Delete pi |
| **device_controller.go** | | | | **Device management** |
| | `/pis/:pi_id/devices` | POST | Admin only | Create device |
| | `/pis/:pi_id/devices` | GET | Admin: all devices<br>User: devices on their PI | List devices |
| | `/pis/:pi_id/devices/:device_id` | GET | Admin: any device<br>User: device on their PI | Get device details |
| | `/pis/:pi_id/devices/:device_id` | PATCH | Admin only | Update device |
| | `/pis/:pi_id/devices/:device_id` | DELETE | Admin only | Delete device |
| **reading_controller.go** | | | | **Reading management** |
| | `/readings/latest?pi_id=X` | GET | Admin: any PI<br>User: their PI only | Get latest readings |
| | `/readings?pi_id=X` | GET | Admin: any PI<br>User: their PI only | Get readings |
| | `/readings/pis/:pi_id/devices/:device_id` | GET | Admin: any device<br>User: device on their PI | Get device readings |
| **health_controller.go** | | | | **Health and stats** |
| | `/health/live` | GET | Public | Liveness check |
| | `/health/ready` | GET | Public | Readiness check |
| | `/metrics` | GET | Public | Metrics endpoint |
| | `/stats/summary` | GET | Admin: all stats<br>User: stats for their resources only | System statistics |

---
