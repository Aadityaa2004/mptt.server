import { apiFetch } from "./apiClient";
import type {
  UserProfile,
  LocationUpdateRequest,
  CurrentWeather,
  WeatherForecast,
  SapFlowDayResponse,
} from "@/types/weather";

export const weatherService = {
  async getProfile(): Promise<UserProfile | null> {
    try {
      const response = await apiFetch("/api/auth/profile", {
        method: "GET",
      });

      if (!response.ok) {
        let errorMessage = "Failed to fetch profile";
        try {
          const error = await response.json();
          errorMessage = error.message || error.detail || errorMessage;
        } catch {
          errorMessage = response.statusText || errorMessage;
        }
        
        // For 404 or 401, return null instead of throwing
        if (response.status === 404 || response.status === 401) {
          return null;
        }
        
        throw new Error(errorMessage);
      }

      const profile = await response.json();
      return profile || null;
    } catch (error) {
      // If it's a network error, log and return null
      if (error instanceof TypeError && error.message.includes("fetch")) {
        console.warn("Network error fetching profile:", error);
        return null;
      }
      throw error;
    }
  },

  async updateProfile(updates: {
    sap_alert_threshold_percent?: number | null;
  }): Promise<UserProfile> {
    const response = await apiFetch("/api/auth/profile", {
      method: "PATCH",
      body: JSON.stringify(updates),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: "Failed to update profile" }));
      throw new Error(error.message || "Failed to update profile");
    }

    return response.json();
  },

  async updateLocation(location: LocationUpdateRequest): Promise<UserProfile> {
    const response = await apiFetch("/api/auth/location", {
      method: "PUT",
      body: JSON.stringify(location),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: "Failed to update location" }));
      throw new Error(error.message || "Failed to update location");
    }

    return response.json();
  },

  async getCurrentWeather(): Promise<CurrentWeather> {
    const response = await apiFetch("/api/auth/weather", {
      method: "GET",
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: "Failed to fetch weather" }));
      throw new Error(error.message || "Failed to fetch weather");
    }

    return response.json();
  },

  async getForecast(): Promise<WeatherForecast> {
    const response = await apiFetch("/api/auth/weather/forecast", {
      method: "GET",
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: "Failed to fetch forecast" }));
      throw new Error(error.message || "Failed to fetch forecast");
    }

    return response.json();
  },

  async getSapFlowDay(): Promise<SapFlowDayResponse> {
    const response = await apiFetch("/api/auth/weather/sap-flow", {
      method: "GET",
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: "Failed to fetch sap flow day" }));
      throw new Error(
        error.error || error.message || error.detail || "Failed to fetch sap flow day"
      );
    }

    return response.json();
  },
};

