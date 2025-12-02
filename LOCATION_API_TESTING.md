# Device Location API Testing Guide

This guide explains how to test the new Device Location CRUD endpoints using Postman.

## Base URL
```
http://localhost:9002
```

## Authentication
All location endpoints require authentication. You'll need to:
1. First, login to get an access token
2. Use the access token in the `Authorization` header for all location requests

---

## Step 1: Login to Get Access Token

**Endpoint:** `POST /api/auth/login`

**Headers:**
```
Content-Type: application/json
```

**Body (JSON):**
```json
{
  "username": "admin",
  "password": "adminpassword123"
}
```

**Response:**
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

**Copy the `access_token` value for use in subsequent requests.**

---

## Step 2: Get All Locations (GET)

**Endpoint:** `GET /api/locations`

**Headers:**
```
Authorization: Bearer <your_access_token>
Content-Type: application/json
```

**Response:**
```json
{
  "locations": []
}
```

Initially, this will return an empty array. After adding locations, it will return all locations for the authenticated user.

---

## Step 3: Add a New Location (POST)

**Endpoint:** `POST /api/locations`

**Headers:**
```
Authorization: Bearer <your_access_token>
Content-Type: application/json
```

**Body (JSON):**
```json
{
  "device_id": "AA:BB:CC:DD:EE:FF",
  "pi_id": "pi_001",
  "device_name": "Temperature Sensor 1",
  "latitude": 40.7128,
  "longitude": -74.0060
}
```

**Note:** `device_name` is optional. You can omit it:
```json
{
  "device_id": "11:22:33:44:55:66",
  "pi_id": "pi_002",
  "latitude": 34.0522,
  "longitude": -118.2437
}
```

**Response (201 Created):**
```json
{
  "device_id": "AA:BB:CC:DD:EE:FF",
  "pi_id": "pi_001",
  "device_name": "Temperature Sensor 1",
  "latitude": 40.7128,
  "longitude": -74.0060
}
```

**Error Cases:**
- **400 Bad Request:** Missing required fields (device_id, pi_id, latitude, longitude)
- **409 Conflict:** Device location with the same device_id already exists

---

## Step 4: Update a Location (PUT)

**Endpoint:** `PUT /api/locations/:device_id`

**Example:** `PUT /api/locations/AA:BB:CC:DD:EE:FF`

**Headers:**
```
Authorization: Bearer <your_access_token>
Content-Type: application/json
```

**Body (JSON):**
```json
{
  "device_id": "AA:BB:CC:DD:EE:FF",
  "pi_id": "pi_001",
  "device_name": "Updated Temperature Sensor 1",
  "latitude": 40.7580,
  "longitude": -73.9855
}
```

**Note:** The `device_id` in the URL path must match the `device_id` in the body.

**Response (200 OK):**
```json
{
  "device_id": "AA:BB:CC:DD:EE:FF",
  "pi_id": "pi_001",
  "device_name": "Updated Temperature Sensor 1",
  "latitude": 40.7580,
  "longitude": -73.9855
}
```

**Error Cases:**
- **400 Bad Request:** Missing required fields or device_id mismatch
- **404 Not Found:** Location with the specified device_id not found

---

## Step 5: Delete a Location (DELETE)

**Endpoint:** `DELETE /api/locations/:device_id`

**Example:** `DELETE /api/locations/AA:BB:CC:DD:EE:FF`

**Headers:**
```
Authorization: Bearer <your_access_token>
Content-Type: application/json
```

**Response (200 OK):**
```json
{
  "message": "location deleted successfully"
}
```

**Error Cases:**
- **404 Not Found:** Location with the specified device_id not found

---

## Complete Testing Flow

### Test Scenario 1: Full CRUD Cycle

1. **Login** → Get access token
2. **GET /api/locations** → Should return empty array `[]`
3. **POST /api/locations** → Add first location
4. **GET /api/locations** → Should return array with 1 location
5. **POST /api/locations** → Add second location
6. **GET /api/locations** → Should return array with 2 locations
7. **PUT /api/locations/:device_id** → Update first location
8. **GET /api/locations** → Verify update
9. **DELETE /api/locations/:device_id** → Delete first location
10. **GET /api/locations** → Should return array with 1 location

### Test Scenario 2: Error Handling

1. **POST /api/locations** with missing `latitude` → Should return 400
2. **POST /api/locations** with duplicate `device_id` → Should return 409
3. **PUT /api/locations/non-existent-id** → Should return 404
4. **DELETE /api/locations/non-existent-id** → Should return 404

---

## Postman Collection Setup

### Environment Variables
Create a Postman environment with:
- `base_url`: `http://localhost:9002`
- `access_token`: (will be set after login)

### Pre-request Script (for Login)
You can create a pre-request script to automatically login and set the token:

```javascript
// Only run if token is not set
if (!pm.environment.get("access_token")) {
    pm.sendRequest({
        url: pm.environment.get("base_url") + "/api/auth/login",
        method: 'POST',
        header: {
            'Content-Type': 'application/json'
        },
        body: {
            mode: 'raw',
            raw: JSON.stringify({
                username: "admin",
                password: "adminpassword123"
            })
        }
    }, function (err, res) {
        if (res.code === 200) {
            var jsonData = res.json();
            pm.environment.set("access_token", jsonData.access_token);
        }
    });
}
```

### Collection Variables
- `{{base_url}}`: `http://localhost:9002`
- `{{access_token}}`: Set from login response

### Request Headers Template
```
Authorization: Bearer {{access_token}}
Content-Type: application/json
```

---

## Example Postman Requests

### 1. Login Request
- **Method:** POST
- **URL:** `{{base_url}}/api/auth/login`
- **Body (raw JSON):**
```json
{
  "username": "admin",
  "password": "adminpassword123"
}
```
- **Tests (to save token):**
```javascript
if (pm.response.code === 200) {
    var jsonData = pm.response.json();
    pm.environment.set("access_token", jsonData.access_token);
}
```

### 2. Get All Locations
- **Method:** GET
- **URL:** `{{base_url}}/api/locations`
- **Headers:** `Authorization: Bearer {{access_token}}`

### 3. Add Location
- **Method:** POST
- **URL:** `{{base_url}}/api/locations`
- **Headers:** `Authorization: Bearer {{access_token}}`
- **Body (raw JSON):**
```json
{
  "device_id": "AA:BB:CC:DD:EE:FF",
  "pi_id": "pi_001",
  "device_name": "Temperature Sensor 1",
  "latitude": 40.7128,
  "longitude": -74.0060
}
```

### 4. Update Location
- **Method:** PUT
- **URL:** `{{base_url}}/api/locations/AA:BB:CC:DD:EE:FF`
- **Headers:** `Authorization: Bearer {{access_token}}`
- **Body (raw JSON):**
```json
{
  "device_id": "AA:BB:CC:DD:EE:FF",
  "pi_id": "pi_001",
  "device_name": "Updated Name",
  "latitude": 40.7580,
  "longitude": -73.9855
}
```

### 5. Delete Location
- **Method:** DELETE
- **URL:** `{{base_url}}/api/locations/AA:BB:CC:DD:EE:FF`
- **Headers:** `Authorization: Bearer {{access_token}}`

---

## Notes

1. **User Isolation:** Each user can only see and manage their own locations. The authenticated user's ID is automatically used.

2. **Device ID Format:** The `device_id` should be a MAC address format (e.g., "AA:BB:CC:DD:EE:FF"), but the API accepts any string format.

3. **Device Name:** The `device_name` field is optional. If not provided, it will be `null` in the response.

4. **Coordinates:** `latitude` and `longitude` are required and should be valid floating-point numbers.

5. **Token Expiration:** Access tokens expire after 24 hours (as configured). If you get a 401 Unauthorized error, login again to get a new token.

---

## Troubleshooting

### 401 Unauthorized
- Make sure you're including the `Authorization: Bearer <token>` header
- Token may have expired - login again
- Check that the token is correctly copied (no extra spaces)

### 404 Not Found
- Verify the `device_id` exists in your locations
- Check that you're using the correct user account (locations are user-specific)

### 500 Internal Server Error
- Check that the database is running: `docker ps`
- Check API service logs: `docker logs api-service`
- Verify the database schema was created correctly

### Database Connection Issues
- Ensure PostgreSQL container is running: `docker ps | grep postgresql`
- Check database logs: `docker logs postgresql`
- Verify environment variables in `docker-compose.yml`

