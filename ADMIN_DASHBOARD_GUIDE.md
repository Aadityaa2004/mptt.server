# Admin Dashboard Implementation Guide

This document provides comprehensive guidelines for building the frontend admin dashboard, including all available admin functionality, API endpoints, request/response formats, and implementation flows.

## Table of Contents

1. [Admin Role Overview](#admin-role-overview)
2. [Authentication Flow](#authentication-flow)
3. [API Base Configuration](#api-base-configuration)
4. [Admin Endpoints Reference](#admin-endpoints-reference)
5. [Data Models](#data-models)
6. [Implementation Workflows](#implementation-workflows)
7. [Error Handling](#error-handling)

---

## Admin Role Overview

### Admin Permissions

The **admin** role has full CRUD (Create, Read, Update, Delete) access to all resources in the system:

- ✅ **Full User Management**: Create, read, update, delete any user
- ✅ **Role Management**: Assign and change user roles (admin/user)
- ✅ **PI Management**: Create, update, delete PIs and assign them to users
- ✅ **Device Management**: Create, update, delete devices on any PI
- ✅ **Reading Access**: View all readings from all devices
- ✅ **Statistics**: View system-wide statistics
- ✅ **Admin Registration**: Create new admin users

### Admin vs User Access

| Resource | Admin Access | User Access |
|----------|-------------|-------------|
| Users | All users | Own profile only |
| PIs | All PIs | Only assigned PIs |
| Devices | All devices | Devices on assigned PIs |
| Readings | All readings | Readings from assigned devices |
| Statistics | System-wide | Own resources only |

---

## Authentication Flow

### Base URL

```
Development: http://localhost:9002
Production: <configured-api-url>
```

### Authentication Headers

All authenticated requests require the access token in the Authorization header:

```
Authorization: Bearer <access_token>
Content-Type: application/json
```

### Login Flow

**Endpoint:** `POST /api/auth/login`

**Request:**
```json
{
  "username": "admin",
  "password": "adminpassword123"
}
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_id": "550e8400-e29b-41d4-a716-446655440000",
  "expires_at": 1761176384,
  "user_id": "user-uuid-here",
  "username": "admin",
  "email": "admin@example.com",
  "role": "admin"
}
```

**Error Responses:**
- `401 Unauthorized`: Invalid credentials
  ```json
  {
    "error": "invalid credentials"
  }
  ```

**Implementation Notes:**
- Store the `token` value for all subsequent API calls
- The `role` field indicates if the user is an admin
- Token expiration is handled via refresh token (stored in HTTP-only cookie)

### Token Refresh

**Endpoint:** `POST /api/auth/refresh`

**Request:** No body required (refresh token sent as HTTP-only cookie)

**Response (200 OK):**
```json
{
  "token": "new-access-token-here",
  "token_id": "new-token-id",
  "expires_at": 1761176384,
  "user_id": "user-uuid-here",
  "username": "admin",
  "email": "admin@example.com",
  "role": "admin"
}
```

### Logout

**Endpoint:** `POST /api/auth/logout`

**Request:** No body required

**Response (200 OK):**
```json
{
  "message": "Logged out successfully"
}
```

---

## Admin Endpoints Reference

### 1. User Management

#### 1.1 Get All Users

**Endpoint:** `GET /api/users`

**Authorization:** Admin only

**Query Parameters:**
- None

**Response (200 OK):**
```json
{
  "users": [
    {
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin",
      "active": true,
      "latitude": 40.7128,
      "longitude": -74.0060,
      "locations": [],
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    },
    {
      "user_id": "660e8400-e29b-41d4-a716-446655440001",
      "username": "user1",
      "email": "user1@example.com",
      "role": "user",
      "active": true,
      "latitude": null,
      "longitude": null,
      "locations": [],
      "created_at": "2024-01-16T08:15:00Z",
      "updated_at": "2024-01-16T08:15:00Z"
    }
  ]
}
```

**Error Responses:**
- `401 Unauthorized`: Not authenticated
- `403 Forbidden`: Not an admin user

---

#### 1.2 Get User by ID

**Endpoint:** `GET /api/users/:id`

**Authorization:** Admin (can access any user) or User (own profile only)

**Path Parameters:**
- `id` (string, required): User ID

**Response (200 OK):**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "admin",
  "email": "admin@example.com",
  "role": "admin",
  "active": true,
  "latitude": 40.7128,
  "longitude": -74.0060,
  "locations": [],
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `404 Not Found`: User not found
- `403 Forbidden`: Access denied (non-admin trying to access another user)

---

#### 1.3 Update User

**Endpoint:** `PUT /api/users/:id`

**Authorization:** Admin only

**Path Parameters:**
- `id` (string, required): User ID

**Request Body:**
```json
{
  "username": "updated_username",
  "email": "updated@example.com",
  "password": "newpassword123",
  "latitude": 45.5017,
  "longitude": -73.5673
}
```

**Note:** All fields are optional. Only include fields you want to update.

**Response (200 OK):**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "updated_username",
  "email": "updated@example.com",
  "role": "admin",
  "active": true,
  "latitude": 45.5017,
  "longitude": -73.5673,
  "locations": [],
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-16T14:20:00Z"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid request body
- `404 Not Found`: User not found

---

#### 1.4 Delete User

**Endpoint:** `DELETE /api/users/:id`

**Authorization:** Admin only

**Path Parameters:**
- `id` (string, required): User ID

**Response (200 OK):**
```json
{
  "message": "user deleted successfully"
}
```

**Error Responses:**
- `404 Not Found`: User not found
- `500 Internal Server Error`: Deletion failed

---

#### 1.5 Update User Role

**Endpoint:** `PUT /api/users/:id/role`

**Authorization:** Admin only

**Path Parameters:**
- `id` (string, required): User ID

**Request Body:**
```json
{
  "role": "admin"
}
```

**Valid Roles:**
- `"admin"`: Administrator with full access
- `"user"`: Regular user with read-only access to assigned resources

**Response (200 OK):**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "user1",
  "email": "user1@example.com",
  "role": "admin",
  "active": true,
  "latitude": null,
  "longitude": null,
  "locations": [],
  "created_at": "2024-01-16T08:15:00Z",
  "updated_at": "2024-01-16T15:30:00Z"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid role or missing role field
- `404 Not Found`: User not found

---

#### 1.6 Register Admin User

**Endpoint:** `POST /api/auth/register/admin`

**Authorization:** Admin only

**Request Body:**
```json
{
  "username": "newadmin",
  "email": "newadmin@example.com",
  "password": "securepassword123"
}
```

**Response (201 Created):**
```json
{
  "id": "770e8400-e29b-41d4-a716-446655440002",
  "username": "newadmin",
  "email": "newadmin@example.com",
  "role": "admin"
}
```

**Error Responses:**
- `400 Bad Request`: Username/email already exists or validation failed
- `403 Forbidden`: Not an admin user

---

### 2. PI Management

#### 2.1 Create PI

**Endpoint:** `POST /pis`

**Authorization:** Admin only

**Request Body:**
```json
{
  "pi_id": "pi_001",
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Note:** `user_id` is optional. If not provided, the PI will be created without assignment.

**Response (201 Created):**
```json
{
  "pi_id": "pi_001",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "created_at": "2024-01-16T16:00:00Z"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid request body or user not found
- `500 Internal Server Error`: Creation failed

---

#### 2.2 List All PIs

**Endpoint:** `GET /pis`

**Authorization:** Admin (all PIs) or User (assigned PIs only)

**Query Parameters:**
- `user_id` (string, optional): Filter by user ID (admin only)
- `page` (int, optional, default: 1): Page number
- `page_size` (int, optional, default: 10): Items per page

**Response (200 OK):**
```json
{
  "items": [
    {
      "pi_id": "pi_001",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "created_at": "2024-01-16T16:00:00Z"
    },
    {
      "pi_id": "pi_002",
      "user_id": "660e8400-e29b-41d4-a716-446655440001",
      "created_at": "2024-01-17T09:30:00Z"
    }
  ],
  "total": 2,
  "page": 1,
  "page_size": 10
}
```

---

#### 2.3 Get PI by ID

**Endpoint:** `GET /pis/:pi_id`

**Authorization:** Admin (any PI) or User (assigned PI only)

**Path Parameters:**
- `pi_id` (string, required): PI ID

**Response (200 OK):**
```json
{
  "pi_id": "pi_001",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "created_at": "2024-01-16T16:00:00Z"
}
```

**Error Responses:**
- `404 Not Found`: PI not found
- `403 Forbidden`: Access denied (non-admin trying to access unassigned PI)

---

#### 2.4 Update PI

**Endpoint:** `PATCH /pis/:pi_id`

**Authorization:** Admin only

**Path Parameters:**
- `pi_id` (string, required): PI ID

**Request Body:**
```json
{
  "user_id": "660e8400-e29b-41d4-a716-446655440001"
}
```

**Response (200 OK):**
```json
{
  "pi_id": "pi_001",
  "user_id": "660e8400-e29b-41d4-a716-446655440001",
  "created_at": "2024-01-16T16:00:00Z"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid request body or user not found
- `404 Not Found`: PI not found

---

#### 2.5 Delete PI

**Endpoint:** `DELETE /pis/:pi_id`

**Authorization:** Admin only

**Path Parameters:**
- `pi_id` (string, required): PI ID

**Query Parameters:**
- `cascade` (boolean, optional, default: false): If true, also deletes associated devices and readings

**Response (200 OK):**
```json
{
  "deleted": true
}
```

**Error Responses:**
- `404 Not Found`: PI not found
- `500 Internal Server Error`: Deletion failed

---

### 3. Device Management

#### 3.1 Create Device

**Endpoint:** `POST /pis/:pi_id/devices`

**Authorization:** Admin only

**Path Parameters:**
- `pi_id` (string, required): PI ID

**Request Body:**
```json
{
  "device_id": 1
}
```

**Response (201 Created):**
```json
{
  "pi_id": "pi_001",
  "device_id": 1
}
```

**Error Responses:**
- `400 Bad Request`: Invalid request body
- `404 Not Found`: PI not found
- `500 Internal Server Error`: Creation failed

---

#### 3.2 List Devices for a PI

**Endpoint:** `GET /pis/:pi_id/devices`

**Authorization:** Admin (any PI) or User (assigned PI only)

**Path Parameters:**
- `pi_id` (string, required): PI ID

**Query Parameters:**
- `page` (int, optional, default: 1): Page number
- `page_size` (int, optional, default: 10): Items per page

**Response (200 OK):**
```json
{
  "items": [
    {
      "pi_id": "pi_001",
      "device_id": 1
    },
    {
      "pi_id": "pi_001",
      "device_id": 2
    }
  ],
  "total": 2,
  "page": 1,
  "page_size": 10
}
```

**Error Responses:**
- `404 Not Found`: PI not found
- `403 Forbidden`: Access denied (non-admin trying to access unassigned PI)

---

#### 3.3 Get Device by ID

**Endpoint:** `GET /pis/:pi_id/devices/:device_id`

**Authorization:** Admin (any device) or User (device on assigned PI only)

**Path Parameters:**
- `pi_id` (string, required): PI ID
- `device_id` (string, required): Device ID

**Response (200 OK):**
```json
{
  "pi_id": "pi_001",
  "device_id": 1
}
```

**Error Responses:**
- `400 Bad Request`: Invalid device_id format
- `404 Not Found`: Device not found
- `403 Forbidden`: Access denied

---

#### 3.4 Update Device

**Endpoint:** `PATCH /pis/:pi_id/devices/:device_id`

**Authorization:** Admin only

**Path Parameters:**
- `pi_id` (string, required): PI ID
- `device_id` (string, required): Device ID

**Request Body:** _No body; no updatable fields for devices currently._

**Response (200 OK):**
```json
{
  "pi_id": "pi_001",
  "device_id": 1
}
```

**Error Responses:**
- `400 Bad Request`: Invalid request body or device_id format
- `404 Not Found`: Device not found

---

#### 3.5 Delete Device

**Endpoint:** `DELETE /pis/:pi_id/devices/:device_id`

**Authorization:** Admin only

**Path Parameters:**
- `pi_id` (string, required): PI ID
- `device_id` (string, required): Device ID

**Query Parameters:**
- `cascade` (boolean, optional, default: false): If true, also deletes associated readings

**Response (200 OK):**
```json
{
  "deleted": true
}
```

**Error Responses:**
- `400 Bad Request`: Invalid device_id format
- `404 Not Found`: Device not found
- `500 Internal Server Error`: Deletion failed

---

### 4. Reading Management

#### 4.1 Get Latest Readings

**Endpoint:** `GET /readings/latest`

**Authorization:** Admin (any PI) or User (assigned PI only)

**Query Parameters:**
- `pi_id` (string, required): PI ID

**Response (200 OK):**
```json
{
  "items": [
    {
      "pi_id": "pi_001",
      "device_id": 1,
      "ts": "2024-01-16T17:00:00Z",
      "payload": {
        "device_id": "AA:BB:CC:DD:EE:FF",
        "pi_id": "pi_001",
        "timestamp": "2024-01-16T17:00:00Z",
        "sensors": {
          "temperature": {
            "value": 25.5,
            "unit": "C"
          }
        },
        "battery_percentage": 85.0
      }
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request`: pi_id is required
- `404 Not Found`: PI not found
- `403 Forbidden`: Access denied

---

#### 4.2 Get Readings

**Endpoint:** `GET /readings`

**Authorization:** Admin (any PI) or User (assigned PI only)

**Query Parameters:**
- `pi_id` (string, required): PI ID
- `device_id` (string, optional): Filter by device ID
- `from` (string, optional): Start time (RFC3339 format, e.g., "2024-01-16T00:00:00Z")
- `to` (string, optional): End time (RFC3339 format, e.g., "2024-01-16T23:59:59Z")
- `limit` (int, optional, default: 100): Maximum number of readings
- `page` (int, optional, default: 1): Page number

**Response (200 OK):**
```json
{
  "items": [
    {
      "pi_id": "pi_001",
      "device_id": 1,
      "ts": "2024-01-16T17:00:00Z",
      "payload": {
        "device_id": "AA:BB:CC:DD:EE:FF",
        "pi_id": "pi_001",
        "timestamp": "2024-01-16T17:00:00Z",
        "sensors": {
          "temperature": {
            "value": 25.5,
            "unit": "C"
          },
          "level": {
            "value": 75.0,
            "unit": "%"
          }
        },
        "battery_percentage": 85.0
      }
    }
  ],
  "total": 150,
  "page": 1,
  "limit": 100
}
```

**Error Responses:**
- `400 Bad Request`: pi_id is required or invalid date format
- `404 Not Found`: PI not found
- `403 Forbidden`: Access denied

---

#### 4.3 Get Device Readings

**Endpoint:** `GET /readings/pis/:pi_id/devices/:device_id`

**Authorization:** Admin (any device) or User (device on assigned PI only)

**Path Parameters:**
- `pi_id` (string, required): PI ID
- `device_id` (int, required): Device ID

**Query Parameters:**
- `from` (string, optional): Start time (RFC3339 format)
- `to` (string, optional): End time (RFC3339 format)
- `limit` (int, optional, default: 100): Maximum number of readings
- `page` (int, optional, default: 1): Page number

**Response (200 OK):**
```json
{
  "items": [
    {
      "pi_id": "pi_001",
      "device_id": 1,
      "ts": "2024-01-16T17:00:00Z",
      "payload": {
        "device_id": "AA:BB:CC:DD:EE:FF",
        "pi_id": "pi_001",
        "timestamp": "2024-01-16T17:00:00Z",
        "sensors": {
          "temperature": {
            "value": 25.5,
            "unit": "C"
          }
        },
        "battery_percentage": 85.0
      }
    }
  ],
  "total": 50,
  "page": 1,
  "limit": 100
}
```

**Error Responses:**
- `400 Bad Request`: Invalid device_id format or invalid date format
- `404 Not Found`: Device or PI not found
- `403 Forbidden`: Access denied

---

### 5. Statistics

#### 5.1 Get Summary Statistics

**Endpoint:** `GET /stats/summary`

**Authorization:** Admin (system-wide) or User (own resources only)

**Query Parameters:**
- `pi_id` (string, optional): Filter by PI ID (required for non-admin users)
- `device_id` (string, optional): Filter by device ID
- `from` (string, optional): Start time (RFC3339 format)
- `to` (string, optional): End time (RFC3339 format)

**Response (200 OK):**
```json
{
  "total_readings": 1250,
  "avg_temperature": 23.5,
  "avg_humidity": 65.2,
  "min_temperature": 18.0,
  "max_temperature": 28.5,
  "date_range": {
    "from": "2024-01-16T00:00:00Z",
    "to": "2024-01-16T23:59:59Z"
  }
}
```

**Error Responses:**
- `400 Bad Request`: pi_id required for non-admin users or invalid date format
- `404 Not Found`: PI not found
- `403 Forbidden`: Access denied

---

### 6. Health & Monitoring

#### 6.1 Health Check (Live)

**Endpoint:** `GET /health/live`

**Authorization:** Public (no authentication required)

**Response (200 OK):**
```json
{
  "status": "ok"
}
```

---

#### 6.2 Health Check (Ready)

**Endpoint:** `GET /health/ready`

**Authorization:** Public (no authentication required)

**Response (200 OK):**
```json
{
  "status": "ready",
  "db": true,
  "mqtt": true
}
```

---

#### 6.3 Metrics

**Endpoint:** `GET /metrics`

**Authorization:** Public (no authentication required)

**Response (200 OK):**
```
# HELP mqtt_ingestor_health Health status of MQTT ingestor
# TYPE mqtt_ingestor_health gauge
mqtt_ingestor_health 1
```

---

## Data Models

### User Model

```typescript
interface User {
  user_id: string;
  username: string;
  email: string;
  role: "admin" | "user";
  active: boolean;
  latitude?: number | null;
  longitude?: number | null;
  locations: DeviceLocation[];
  created_at: string; // ISO 8601 format
  updated_at: string; // ISO 8601 format
}
```

### PI Model

```typescript
interface Pi {
  pi_id: string;
  user_id?: string | null;
  created_at: string; // ISO 8601 format
}
```

### Device Model

```typescript
interface Device {
  pi_id: string;
  device_id: number;
  device_type: "temperature" | "humidity" | "light" | "pressure";
  created_at: string; // ISO 8601 format
}
```

### Reading Model

```typescript
interface Reading {
  pi_id: string;
  device_id: number;
  ts: string; // ISO 8601 format
  payload: ReadingPayload;
}

interface ReadingPayload {
  device_id: string; // MAC address
  pi_id: string;
  timestamp: string; // ISO 8601 format
  sensors: {
    temperature?: {
      value: number;
      unit: string;
    };
    level?: {
      value: number;
      unit: string;
    };
  };
  battery_percentage: number;
}
```

### Device Location Model

```typescript
interface DeviceLocation {
  device_id: string; // MAC address
  pi_id: string;
  device_name?: string;
  latitude: number;
  longitude: number;
}
```

### Paginated Response Model

```typescript
interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size?: number;
  limit?: number;
}
```

---

## Implementation Workflows

### Workflow 1: Admin Login and Dashboard Initialization

1. **Login**
   - POST `/api/auth/login` with admin credentials
   - Store access token in memory/localStorage
   - Store user role to determine admin access

2. **Load Dashboard Data**
   - Parallel API calls:
     - GET `/api/users` - Load all users
     - GET `/pis` - Load all PIs
     - GET `/stats/summary` - Load system statistics

3. **Display Dashboard**
   - Show user count, PI count, device count
   - Display recent activity
   - Show system health status

---

### Workflow 2: Create and Assign PI to User

1. **Get User List**
   - GET `/api/users` to populate user dropdown

2. **Create PI**
   - POST `/pis` with `pi_id` and optional `user_id`

3. **Update PI Assignment (if needed)**
   - PATCH `/pis/:pi_id` to change `user_id`

4. **Refresh PI List**
   - GET `/pis` to show updated list

---

### Workflow 3: Create Device on PI

1. **Select PI**
   - User selects PI from list (GET `/pis`)

2. **Create Device**
   - POST `/pis/:pi_id/devices` with `device_id` and `device_type`

3. **Verify Creation**
   - GET `/pis/:pi_id/devices/:device_id` to confirm

4. **Refresh Device List**
   - GET `/pis/:pi_id/devices` to show updated list

---

### Workflow 4: Manage User Roles

1. **View Users**
   - GET `/api/users` to see all users with their roles

2. **Update Role**
   - PUT `/api/users/:id/role` with new role

3. **Verify Change**
   - GET `/api/users/:id` to confirm role update

---

### Workflow 5: View Readings and Statistics

1. **Select PI**
   - Admin selects PI from dropdown (GET `/pis`)

2. **View Latest Readings**
   - GET `/readings/latest?pi_id=:pi_id`

3. **View Historical Readings**
   - GET `/readings?pi_id=:pi_id&from=:from&to=:to`

4. **View Statistics**
   - GET `/stats/summary?pi_id=:pi_id&from=:from&to=:to`

---

### Workflow 6: Delete Resources (Cascade)

1. **Delete Device**
   - DELETE `/pis/:pi_id/devices/:device_id?cascade=true`
   - This deletes the device and all its readings

2. **Delete PI**
   - DELETE `/pis/:pi_id?cascade=true`
   - This deletes the PI, all its devices, and all readings

3. **Delete User**
   - DELETE `/api/users/:id`
   - Note: This is a hard delete. Consider soft delete or reassignment of resources first.

---

## Error Handling

### Common Error Responses

#### 401 Unauthorized
```json
{
  "error": "Authentication required"
}
```
**Action:** Redirect to login page

#### 403 Forbidden
```json
{
  "error": "Admin access required"
}
```
**Action:** Show error message, hide admin-only features

#### 404 Not Found
```json
{
  "error": "resource not found"
}
```
**Action:** Show "Resource not found" message

#### 400 Bad Request
```json
{
  "error": "Invalid request: <details>"
}
```
**Action:** Show validation error message

#### 500 Internal Server Error
```json
{
  "error": "Internal server error"
}
```
**Action:** Show generic error, log for debugging

### Error Handling Best Practices

1. **Token Expiration**
   - If receiving 401, attempt token refresh
   - If refresh fails, redirect to login

2. **Network Errors**
   - Show retry option
   - Implement exponential backoff

3. **Validation Errors**
   - Display field-specific errors
   - Highlight invalid fields in forms

4. **Permission Errors**
   - Check user role before showing admin features
   - Gracefully handle 403 responses

---

## Frontend Implementation Checklist

### Authentication
- [ ] Login page with username/password
- [ ] Token storage (localStorage or memory)
- [ ] Token refresh mechanism
- [ ] Logout functionality
- [ ] Protected route wrapper
- [ ] Role-based access control (RBAC) checks

### User Management
- [ ] User list page with table/grid
- [ ] User detail view
- [ ] Create user form (admin registration)
- [ ] Edit user form
- [ ] Delete user confirmation dialog
- [ ] Role change dropdown/selector

### PI Management
- [ ] PI list page
- [ ] Create PI form
- [ ] Edit PI form (reassign user)
- [ ] Delete PI confirmation (with cascade option)
- [ ] PI detail view

### Device Management
- [ ] Device list per PI
- [ ] Create device form
- [ ] Edit device form (change type)
- [ ] Delete device confirmation (with cascade option)
- [ ] Device detail view

### Reading Management
- [ ] Latest readings view
- [ ] Historical readings view with filters
- [ ] Date range picker
- [ ] Reading charts/graphs
- [ ] Export readings functionality

### Statistics
- [ ] Dashboard with summary statistics
- [ ] Statistics filters (PI, device, date range)
- [ ] Statistics charts/graphs

### UI/UX
- [ ] Responsive design
- [ ] Loading states
- [ ] Error messages
- [ ] Success notifications
- [ ] Confirmation dialogs
- [ ] Pagination controls
- [ ] Search/filter functionality

---

## API Client Example (TypeScript/JavaScript)

```typescript
class AdminAPIClient {
  private baseURL: string;
  private accessToken: string | null = null;

  constructor(baseURL: string = 'http://localhost:9002') {
    this.baseURL = baseURL;
  }

  setAccessToken(token: string) {
    this.accessToken = token;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    if (this.accessToken) {
      headers['Authorization'] = `Bearer ${this.accessToken}`;
    }

    const response = await fetch(url, {
      ...options,
      headers,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    return response.json();
  }

  // Authentication
  async login(username: string, password: string) {
    const response = await this.request<AuthResponse>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    });
    this.setAccessToken(response.token);
    return response;
  }

  // User Management
  async getAllUsers() {
    return this.request<{ users: User[] }>('/api/users');
  }

  async getUserById(id: string) {
    return this.request<User>(`/api/users/${id}`);
  }

  async updateUser(id: string, updates: Partial<User>) {
    return this.request<User>(`/api/users/${id}`, {
      method: 'PUT',
      body: JSON.stringify(updates),
    });
  }

  async deleteUser(id: string) {
    return this.request<{ message: string }>(`/api/users/${id}`, {
      method: 'DELETE',
    });
  }

  async updateUserRole(id: string, role: 'admin' | 'user') {
    return this.request<User>(`/api/users/${id}/role`, {
      method: 'PUT',
      body: JSON.stringify({ role }),
    });
  }

  async registerAdmin(userData: { username: string; email: string; password: string }) {
    return this.request<User>('/api/auth/register/admin', {
      method: 'POST',
      body: JSON.stringify(userData),
    });
  }

  // PI Management
  async createPi(piData: { pi_id: string; user_id?: string }) {
    return this.request<Pi>('/pis', {
      method: 'POST',
      body: JSON.stringify(piData),
    });
  }

  async getAllPis(userId?: string, page = 1, pageSize = 10) {
    const params = new URLSearchParams({
      page: page.toString(),
      page_size: pageSize.toString(),
    });
    if (userId) params.append('user_id', userId);
    return this.request<PaginatedResponse<Pi>>(`/pis?${params}`);
  }

  async getPiById(piId: string) {
    return this.request<Pi>(`/pis/${piId}`);
  }

  async updatePi(piId: string, updates: { user_id?: string }) {
    return this.request<Pi>(`/pis/${piId}`, {
      method: 'PATCH',
      body: JSON.stringify(updates),
    });
  }

  async deletePi(piId: string, cascade = false) {
    const params = cascade ? '?cascade=true' : '';
    return this.request<{ deleted: boolean }>(`/pis/${piId}${params}`, {
      method: 'DELETE',
    });
  }

  // Device Management
  async createDevice(piId: string, deviceData: { device_id: number; device_type: string }) {
    return this.request<Device>(`/pis/${piId}/devices`, {
      method: 'POST',
      body: JSON.stringify(deviceData),
    });
  }

  async getDevices(piId: string, page = 1, pageSize = 10) {
    const params = new URLSearchParams({
      page: page.toString(),
      page_size: pageSize.toString(),
    });
    return this.request<PaginatedResponse<Device>>(`/pis/${piId}/devices?${params}`);
  }

  async getDevice(piId: string, deviceId: number) {
    return this.request<Device>(`/pis/${piId}/devices/${deviceId}`);
  }

  async updateDevice(piId: string, deviceId: number, updates: { device_type?: string }) {
    return this.request<Device>(`/pis/${piId}/devices/${deviceId}`, {
      method: 'PATCH',
      body: JSON.stringify(updates),
    });
  }

  async deleteDevice(piId: string, deviceId: number, cascade = false) {
    const params = cascade ? '?cascade=true' : '';
    return this.request<{ deleted: boolean }>(`/pis/${piId}/devices/${deviceId}${params}`, {
      method: 'DELETE',
    });
  }

  // Reading Management
  async getLatestReadings(piId: string) {
    return this.request<{ items: Reading[] }>(`/readings/latest?pi_id=${piId}`);
  }

  async getReadings(params: {
    pi_id: string;
    device_id?: string;
    from?: string;
    to?: string;
    limit?: number;
    page?: number;
  }) {
    const queryParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined) queryParams.append(key, value.toString());
    });
    return this.request<PaginatedResponse<Reading>>(`/readings?${queryParams}`);
  }

  async getDeviceReadings(piId: string, deviceId: number, params?: {
    from?: string;
    to?: string;
    limit?: number;
    page?: number;
  }) {
    const queryParams = new URLSearchParams();
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) queryParams.append(key, value.toString());
      });
    }
    return this.request<PaginatedResponse<Reading>>(
      `/readings/pis/${piId}/devices/${deviceId}${queryParams.toString() ? '?' + queryParams : ''}`
    );
  }

  // Statistics
  async getSummaryStats(params?: {
    pi_id?: string;
    device_id?: string;
    from?: string;
    to?: string;
  }) {
    const queryParams = new URLSearchParams();
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) queryParams.append(key, value);
      });
    }
    return this.request<any>(`/stats/summary${queryParams.toString() ? '?' + queryParams : ''}`);
  }
}

// Usage example
const api = new AdminAPIClient('http://localhost:9002');

// Login
const authResponse = await api.login('admin', 'adminpassword123');
console.log('Logged in as:', authResponse.username);

// Get all users
const { users } = await api.getAllUsers();
console.log('Total users:', users.length);

// Create a PI
const newPi = await api.createPi({ pi_id: 'pi_003', user_id: users[0].user_id });
console.log('Created PI:', newPi);

// Create a device
const newDevice = await api.createDevice('pi_003', {
  device_id: 1,
  device_type: 'temperature',
});
console.log('Created device:', newDevice);
```

---

## Summary

This guide provides everything needed to build a comprehensive admin dashboard:

1. **Complete API Reference**: All admin endpoints with request/response examples
2. **Data Models**: TypeScript interfaces for all data structures
3. **Workflows**: Step-by-step implementation flows for common operations
4. **Error Handling**: Common errors and how to handle them
5. **Implementation Checklist**: Features to implement
6. **API Client Example**: Ready-to-use TypeScript client

The admin dashboard should provide:
- User management (CRUD operations)
- PI management (create, assign, update, delete)
- Device management (create, update, delete)
- Reading visualization and filtering
- System statistics and monitoring
- Role management

All endpoints require admin authentication except for public health checks. The system uses JWT tokens for authentication, with refresh tokens stored in HTTP-only cookies.

