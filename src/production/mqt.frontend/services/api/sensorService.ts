import { apiFetch, getAccessToken } from "./apiClient";
import { READINGS_API_BASE_URL } from "@/constants/api";
import type { PaginatedResponse, Reading } from "@/types/admin";

export interface Pi {
  pi_id: string;
  user_id: string;
}

export interface PiDevice {
  pi_id: string;
  device_id: string;
}

export interface GetPisParams {
  user_id: string;
  page?: number;
  page_size?: number;
}

export interface GetDevicesParams {
  pi_id: string;
  page?: number;
  page_size?: number;
}

export interface GetReadingsParams {
  page?: number;
  limit?: number;
  from?: string;
  to?: string;
}

export interface ReadingsResponse {
  items: Reading[];
  next_page_token?: string | null;
}

/**
 * Helper function to fetch from readings API (port 8080)
 */
async function readingsApiFetch(url: string, options: RequestInit = {}): Promise<Response> {
  const token = getAccessToken();
  const headers = new Headers(options.headers);
  
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }
  
  if (!headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(`${READINGS_API_BASE_URL}${url}`, {
    ...options,
    headers,
    credentials: "include",
  });

  return response;
}

const sensorService = {
  /**
   * Get all PIs for a user
   */
  async getPis(params: GetPisParams): Promise<PaginatedResponse<Pi>> {
    const { user_id, page = 1, page_size = 10 } = params;
    const queryParams = new URLSearchParams({
      user_id,
      page: page.toString(),
      page_size: page_size.toString(),
    });

    const response = await apiFetch(`/pis?${queryParams.toString()}`, {
      method: "GET",
    });

    if (!response.ok) {
      // For 404 or 401, return empty array (new users may not have PIs yet)
      if (response.status === 404 || response.status === 401) {
        return { items: [] };
      }
      const error = await response.json().catch(() => ({ message: "Failed to fetch PIs" }));
      throw new Error(error.message || "Failed to fetch PIs");
    }

    const data = await response.json();
    
    // Ensure we always return a valid response structure
    if (!data || typeof data !== "object") {
      return { items: [] };
    }
    
    return {
      items: Array.isArray(data.items) ? data.items : [],
      total: data.total,
      page: data.page,
      page_size: data.page_size,
      next_page: data.next_page || null,
    };
  },

  /**
   * Get all devices for a PI
   */
  async getDevices(params: GetDevicesParams): Promise<PaginatedResponse<PiDevice>> {
    const { pi_id, page = 1, page_size = 10 } = params;
    const queryParams = new URLSearchParams({
      page: page.toString(),
      page_size: page_size.toString(),
    });

    const response = await apiFetch(`/pis/${pi_id}/devices?${queryParams.toString()}`, {
      method: "GET",
    });

    if (!response.ok) {
      // For 404 or 401, return empty array (PI may not have devices yet)
      if (response.status === 404 || response.status === 401) {
        return { items: [] };
      }
      const error = await response.json().catch(() => ({ message: "Failed to fetch devices" }));
      throw new Error(error.message || "Failed to fetch devices");
    }

    const data = await response.json();
    
    // Ensure we always return a valid response structure
    if (!data || typeof data !== "object") {
      return { items: [] };
    }
    
    return {
      items: Array.isArray(data.items) ? data.items : [],
      total: data.total,
      page: data.page,
      page_size: data.page_size,
      next_page: data.next_page || null,
    };
  },

  /**
   * Get latest readings for all devices on a PI
   */
  async getLatestReadings(piId: string): Promise<ReadingsResponse> {
    const queryParams = new URLSearchParams({
      pi_id: piId,
    });

    const response = await readingsApiFetch(`/readings/latest?${queryParams.toString()}`, {
      method: "GET",
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to fetch latest readings" }));
      throw new Error(error.error || error.message || "Failed to fetch latest readings");
    }

    return response.json();
  },

  /**
   * Get readings with optional filtering
   */
  async getReadings(
    piId: string,
    deviceId?: string,
    params?: GetReadingsParams
  ): Promise<ReadingsResponse> {
    const queryParams = new URLSearchParams({
      pi_id: piId,
    });

    if (deviceId) {
      queryParams.append("device_id", deviceId);
    }

    if (params) {
      if (params.page !== undefined) {
        queryParams.append("page", params.page.toString());
      }
      if (params.limit !== undefined) {
        queryParams.append("limit", params.limit.toString());
      }
      if (params.from) {
        queryParams.append("from", params.from);
      }
      if (params.to) {
        queryParams.append("to", params.to);
      }
    }

    const response = await readingsApiFetch(`/readings?${queryParams.toString()}`, {
      method: "GET",
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to fetch readings" }));
      throw new Error(error.error || error.message || "Failed to fetch readings");
    }

    const data = await response.json();
    
    // Ensure we always return a valid response structure
    if (!data || typeof data !== "object") {
      return { items: [], next_page_token: null };
    }
    
    return {
      items: Array.isArray(data.items) ? data.items : [],
      next_page_token: data.next_page_token || null,
    };
  },

  /**
   * Get device readings (device_id is MAC address string like "AA:BB:CC:DD:EE:FF")
   */
  async getDeviceReadings(
    piId: string,
    deviceId: string,
    params?: GetReadingsParams
  ): Promise<ReadingsResponse> {
    const queryParams = new URLSearchParams();
    
    if (params) {
      if (params.page !== undefined) {
        queryParams.append("page", params.page.toString());
      }
      if (params.limit !== undefined) {
        queryParams.append("limit", params.limit.toString());
      }
      if (params.from) {
        queryParams.append("from", params.from);
      }
      if (params.to) {
        queryParams.append("to", params.to);
      }
    }

    // Use deviceId as-is (MAC address string)
    // URL encode to handle colons in MAC address
    const encodedDeviceId = encodeURIComponent(deviceId);
    const url = `/readings/pis/${piId}/devices/${encodedDeviceId}${queryParams.toString() ? `?${queryParams}` : ""}`;
    
    const response = await readingsApiFetch(url, {
      method: "GET",
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to fetch device readings" }));
      throw new Error(error.error || error.message || "Failed to fetch device readings");
    }

    return response.json();
  },

  /**
   * Get latest reading for a specific device
   */
  async getLatestDeviceReading(
    piId: string,
    deviceId: string
  ): Promise<Reading | null> {
    try {
      const response = await this.getReadings(piId, deviceId, {
        limit: 1,
      });

      if (!response || !response.items || !Array.isArray(response.items)) {
        return null;
      }

      return response.items.length > 0 ? response.items[0] : null;
    } catch (err) {
      console.error("Error fetching latest device reading:", err);
      return null;
    }
  },
};

export { sensorService };

