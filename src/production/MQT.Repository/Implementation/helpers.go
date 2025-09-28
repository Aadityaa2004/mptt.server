package implementation

// ensureMetaNotNull ensures Meta is not nil to prevent null JSON issues
func ensureMetaNotNull(meta map[string]interface{}) map[string]interface{} {
	if meta == nil {
		return make(map[string]interface{})
	}
	return meta
}

// User Repository (CRUD)
// ├── CreateUser() - Idempotent upsert
// ├── GetUser() - Single user lookup
// ├── ListUsers() - All users
// ├── UpdateUser() - Update name/role/meta
// └── DeleteUser() - Remove user

// Pi Repository (CRUD)
// ├── CreatePi() - Idempotent upsert with user_id
// ├── GetPi() - Single Pi lookup
// ├── ListPis() - All Pis
// ├── ListPisByUser() - User's Pis (NEW)
// ├── UpdatePi() - Update user_id/meta
// └── DeletePi() - Remove Pi

// Device Repository (CRUD)
// ├── CreateDevice() - Idempotent upsert
// ├── GetDevice() - Single device lookup
// ├── ListDevicesByPi() - Pi's devices
// ├── ListAllDevices() - All devices
// ├── UpdateDevice() - Update meta
// └── DeleteDevice() - Remove device

// Reading Repository (Time-Series)
// ├── CreateReading() - Single reading upsert
// ├── CreateReadings() - Bulk upsert (FIXED)
// ├── GetReadingsByPi() - Pi's readings
// ├── GetReadingsByDevice() - Device readings
// ├── GetReadingsByTimeRange() - Time-filtered
// ├── GetLatestReadings() - Latest per device
// └── DeleteReadingsByTimeRange() - Cleanup
