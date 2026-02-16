import { apiFetch } from "@/services/api/apiClient";
import type {
  User,
  Pi,
  Device,
  Reading,
  PaginatedResponse,
  SummaryStatistics,
  CreatePiRequest,
  UpdatePiRequest,
  CreateDeviceRequest,
  UpdateDeviceRequest,
  UpdateUserRequest,
  UpdateUserRoleRequest,
  RegisterAdminRequest,
  GetReadingsParams,
  GetDeviceReadingsParams,
  GetStatsParams,
} from "@/types/admin";

// User Management
export const adminService = {
  // User Management
  async getAllUsers(): Promise<{ users: User[] }> {
    const response = await apiFetch("/api/users");
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to fetch users" }));
      throw new Error(error.error || "Failed to fetch users");
    }
    return response.json();
  },

  async getUserById(id: string): Promise<User> {
    const response = await apiFetch(`/api/users/${id}`);
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to fetch user" }));
      throw new Error(error.error || "Failed to fetch user");
    }
    return response.json();
  },

  async updateUser(id: string, updates: UpdateUserRequest): Promise<User> {
    const response = await apiFetch(`/api/users/${id}`, {
      method: "PUT",
      body: JSON.stringify(updates),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to update user" }));
      throw new Error(error.error || "Failed to update user");
    }
    return response.json();
  },

  async deleteUser(id: string): Promise<{ message: string }> {
    const response = await apiFetch(`/api/users/${id}`, {
      method: "DELETE",
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to delete user" }));
      throw new Error(error.error || "Failed to delete user");
    }
    return response.json();
  },

  async updateUserRole(id: string, role: "admin" | "user"): Promise<User> {
    const response = await apiFetch(`/api/users/${id}/role`, {
      method: "PUT",
      body: JSON.stringify({ role }),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to update user role" }));
      throw new Error(error.error || "Failed to update user role");
    }
    return response.json();
  },

  async registerAdmin(userData: RegisterAdminRequest): Promise<User> {
    const response = await apiFetch("/api/auth/register/admin", {
      method: "POST",
      body: JSON.stringify(userData),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to register admin" }));
      throw new Error(error.error || "Failed to register admin");
    }
    return response.json();
  },

  // PI Management
  async createPi(piData: CreatePiRequest): Promise<Pi> {
    const response = await apiFetch("/pis", {
      method: "POST",
      body: JSON.stringify(piData),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to create PI" }));
      throw new Error(error.error || "Failed to create PI");
    }
    return response.json();
  },

  async getAllPis(userId?: string, page = 1, pageSize = 10): Promise<PaginatedResponse<Pi>> {
    const params = new URLSearchParams({
      page: page.toString(),
      page_size: pageSize.toString(),
    });
    if (userId) params.append("user_id", userId);
    const response = await apiFetch(`/pis?${params}`);
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to fetch PIs" }));
      throw new Error(error.error || "Failed to fetch PIs");
    }
    return response.json();
  },

  async getPiById(piId: string): Promise<Pi> {
    const response = await apiFetch(`/pis/${piId}`);
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to fetch PI" }));
      throw new Error(error.error || "Failed to fetch PI");
    }
    return response.json();
  },

  async updatePi(piId: string, updates: UpdatePiRequest): Promise<Pi> {
    const response = await apiFetch(`/pis/${piId}`, {
      method: "PATCH",
      body: JSON.stringify(updates),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to update PI" }));
      throw new Error(error.error || "Failed to update PI");
    }
    return response.json();
  },

  async deletePi(piId: string, cascade = false): Promise<{ deleted: boolean }> {
    const params = cascade ? "?cascade=true" : "";
    const response = await apiFetch(`/pis/${piId}${params}`, {
      method: "DELETE",
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to delete PI" }));
      throw new Error(error.error || "Failed to delete PI");
    }
    return response.json();
  },

  // Device Management
  async createDevice(piId: string, deviceData: CreateDeviceRequest): Promise<Device> {
    // Validate input
    if (deviceData.device_id === undefined || deviceData.device_id === null || deviceData.device_id === "") {
      throw new Error("Device ID is required");
    }
    
    // Always send device_id as a string (backend Go API expects string type)
    // Convert to string if it's a number (e.g., 123 -> "123")
    const deviceIdString = String(deviceData.device_id);
    
    const requestBody: Record<string, unknown> = {
      device_id: deviceIdString,
    };
    
    const response = await apiFetch(`/pis/${piId}/devices`, {
      method: "POST",
      body: JSON.stringify(requestBody),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to create device" }));
      throw new Error(error.error || "Failed to create device");
    }
    return response.json();
  },

  async getDevices(piId: string, page = 1, pageSize = 10): Promise<PaginatedResponse<Device>> {
    const params = new URLSearchParams({
      page: page.toString(),
      page_size: pageSize.toString(),
    });
    const response = await apiFetch(`/pis/${piId}/devices?${params}`);
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to fetch devices" }));
      throw new Error(error.error || "Failed to fetch devices");
    }
    return response.json();
  },

  async getDevice(piId: string, deviceId: number | string): Promise<Device> {
    // URL encode deviceId to handle MAC addresses with colons
    const encodedDeviceId = encodeURIComponent(String(deviceId));
    const response = await apiFetch(`/pis/${piId}/devices/${encodedDeviceId}`);
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to fetch device" }));
      throw new Error(error.error || "Failed to fetch device");
    }
    return response.json();
  },

  async updateDevice(
    piId: string,
    deviceId: number | string,
    updates: UpdateDeviceRequest
  ): Promise<Device> {
    // Device updates not currently supported by backend
    const requestBody: Record<string, unknown> = {};
    // URL encode deviceId to handle MAC addresses with colons
    const encodedDeviceId = encodeURIComponent(String(deviceId));
    const response = await apiFetch(`/pis/${piId}/devices/${encodedDeviceId}`, {
      method: "PATCH",
      body: JSON.stringify(requestBody),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to update device" }));
      throw new Error(error.error || "Failed to update device");
    }
    return response.json();
  },

  async deleteDevice(piId: string, deviceId: number | string, cascade = false): Promise<{ deleted: boolean }> {
    const params = cascade ? "?cascade=true" : "";
    // URL encode deviceId to handle MAC addresses with colons
    const encodedDeviceId = encodeURIComponent(String(deviceId));
    const url = `/pis/${piId}/devices/${encodedDeviceId}${params}`;
    const response = await apiFetch(url, {
      method: "DELETE",
    });
    if (!response.ok) {
      let errorMessage = "Failed to delete device";
      try {
        const errorData = await response.json();
        errorMessage = errorData.error || errorData.message || errorMessage;
      } catch {
        // If response is not JSON, try to get status text
        errorMessage = response.statusText || errorMessage;
      }
      console.error(`Delete device failed: ${url}`, {
        status: response.status,
        statusText: response.statusText,
        deviceId,
        encodedDeviceId,
      });
      throw new Error(errorMessage);
    }
    return response.json();
  },

  // Reading Management
  async getLatestReadings(piId: string): Promise<{ items: Reading[] }> {
    const response = await apiFetch(`/readings/latest?pi_id=${piId}`);
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to fetch latest readings" }));
      throw new Error(error.error || "Failed to fetch latest readings");
    }
    return response.json();
  },

  async getReadings(params: GetReadingsParams): Promise<PaginatedResponse<Reading>> {
    const queryParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined) queryParams.append(key, value.toString());
    });
    const response = await apiFetch(`/readings?${queryParams}`);
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to fetch readings" }));
      throw new Error(error.error || "Failed to fetch readings");
    }
    return response.json();
  },

  async getDeviceReadings(
    piId: string,
    deviceId: number | string,
    params?: GetDeviceReadingsParams
  ): Promise<PaginatedResponse<Reading>> {
    const queryParams = new URLSearchParams();
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) queryParams.append(key, value.toString());
      });
    }
    // URL encode deviceId to handle MAC addresses with colons
    const encodedDeviceId = encodeURIComponent(String(deviceId));
    const url = `/readings/pis/${piId}/devices/${encodedDeviceId}${queryParams.toString() ? `?${queryParams}` : ""}`;
    const response = await apiFetch(url);
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to fetch device readings" }));
      throw new Error(error.error || "Failed to fetch device readings");
    }
    return response.json();
  },

  // Statistics
  async getSummaryStats(params?: GetStatsParams): Promise<SummaryStatistics> {
    const queryParams = new URLSearchParams();
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) queryParams.append(key, value);
      });
    }
    const url = `/stats/summary${queryParams.toString() ? `?${queryParams}` : ""}`;
    const response = await apiFetch(url);
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to fetch statistics" }));
      throw new Error(error.error || "Failed to fetch statistics");
    }
    return response.json();
  },

  // Health Checks
  async getIngestorHealth(): Promise<{
    status: string;
    timestamp: string;
    services: {
      mqtt: string;
      api_service: string;
    };
    circuit_breaker: {
      state: string;
      failure_count: number;
    };
  } | null> {
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 5000);
      
      const response = await fetch("http://localhost:9003/health", {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
        signal: controller.signal,
      });
      
      clearTimeout(timeoutId);
      
      if (!response.ok) {
        return null;
      }
      return response.json();
    } catch (error) {
      // Handle network errors, CORS errors, timeouts, etc.
      console.error("Ingestor health check failed:", error);
      return null;
    }
  },

  async getApiHealth(): Promise<{
    db: boolean;
    mqtt: boolean;
    status: string;
  } | null> {
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 5000);
      
      const response = await fetch("http://localhost:9002/health/ready", {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
        signal: controller.signal,
      });
      
      clearTimeout(timeoutId);
      
      if (!response.ok) {
        return null;
      }
      return response.json();
    } catch (error) {
      // Handle network errors, CORS errors, timeouts, etc.
      console.error("API health check failed:", error);
      return null;
    }
  },
};

