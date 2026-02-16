import { apiFetch } from "./apiClient";
import { API_ENDPOINTS } from "@/constants/api";
import type { Device, CreateDeviceRequest } from "@/types/device";

export interface DeviceLocation {
  device_id: string;
  pi_id: string;
  device_name?: string;
  latitude: number;
  longitude: number;
  color?: string; // Hex color code (e.g., "#FF5733")
}

export interface DeviceLocationsResponse {
  locations: DeviceLocation[];
}

export interface DeleteLocationResponse {
  message: string;
}

const deviceLocationService = {
  /**
   * Get all device locations for the authenticated user
   */
  async getAllLocations(): Promise<DeviceLocation[]> {
    try {
      const response = await apiFetch(API_ENDPOINTS.LOCATIONS.BASE, {
        method: "GET",
      });

      if (!response.ok) {
        let errorMessage = "Failed to fetch locations";
        try {
          const error = await response.json();
          errorMessage = error.message || error.detail || errorMessage;
        } catch {
          // If response is not JSON, use status text or default message
          errorMessage = response.statusText || errorMessage;
        }
        
        // Don't throw error for 404 or 401 - just return empty array (no locations yet)
        // This is normal for new users who haven't added devices yet
        if (response.status === 404 || response.status === 401) {
          return [];
        }
        
        throw new Error(errorMessage);
      }

      const data: DeviceLocationsResponse = await response.json();
      return data.locations || [];
    } catch (error) {
      // If it's a network error or fetch failed, log and return empty array
      if (error instanceof TypeError && error.message.includes("fetch")) {
        console.warn("Network error fetching locations:", error);
        return [];
      }
      // Re-throw other errors
      throw error;
    }
  },

  /**
   * Add a new device location
   */
  async addLocation(location: CreateDeviceRequest & { color?: string }): Promise<DeviceLocation> {
    const response = await apiFetch(API_ENDPOINTS.LOCATIONS.BASE, {
      method: "POST",
      body: JSON.stringify({
        device_id: location.device_id,
        pi_id: location.pi_id,
        latitude: location.latitude,
        longitude: location.longitude,
        ...(location.color && { color: location.color }),
      }),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: "Failed to add location" }));
      throw new Error((error as { message?: string }).message || "Failed to add location");
    }

    return response.json();
  },

  /**
   * Update an existing device location
   */
  async updateLocation(deviceId: string, location: CreateDeviceRequest & { color?: string }): Promise<DeviceLocation> {
    const response = await apiFetch(API_ENDPOINTS.LOCATIONS.BY_DEVICE_ID(deviceId), {
      method: "PUT",
      body: JSON.stringify({
        device_id: location.device_id,
        pi_id: location.pi_id,
        latitude: location.latitude,
        longitude: location.longitude,
        ...(location.color && { color: location.color }),
      }),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: "Failed to update location" }));
      throw new Error((error as { message?: string }).message || "Failed to update location");
    }

    return response.json();
  },

  /**
   * Update color for all device locations belonging to a specific PI
   */
  async updatePiColor(piId: string, color: string): Promise<void> {
    try {
      // Get all locations for this PI
      const allLocations = await this.getAllLocations();
      const piLocations = allLocations.filter(loc => loc.pi_id === piId);

      // Update each location with the new color
      await Promise.all(
        piLocations.map(location =>
          this.updateLocation(location.device_id, {
            device_id: location.device_id,
            pi_id: location.pi_id,
            latitude: location.latitude,
            longitude: location.longitude,
            color: color,
          })
        )
      );
    } catch (error) {
      console.error(`Error updating color for PI ${piId}:`, error);
      throw error;
    }
  },

  /**
   * Delete a device location
   */
  async deleteLocation(deviceId: string): Promise<void> {
    const response = await apiFetch(API_ENDPOINTS.LOCATIONS.BY_DEVICE_ID(deviceId), {
      method: "DELETE",
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: "Failed to delete location" }));
      throw new Error((error as { message?: string }).message || "Failed to delete location");
    }

    await response.json(); // Parse the response but don't use it
  },

  /**
   * Convert DeviceLocation to Device format
   */
  convertToDevice(location: DeviceLocation, id?: string): Device {
    return {
      id: id || location.device_id,
      device_id: location.device_id,
      pi_id: location.pi_id,
      name: location.device_name,
      latitude: location.latitude,
      longitude: location.longitude,
      color: location.color,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
  },
};

export { deviceLocationService };

