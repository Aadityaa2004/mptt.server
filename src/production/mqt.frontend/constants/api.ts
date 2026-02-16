export const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:9002";
export const READINGS_API_BASE_URL = process.env.NEXT_PUBLIC_READINGS_API_BASE_URL || "http://localhost:9002";

export const API_ENDPOINTS = {
  AUTH: {
    LOGIN: "/api/auth/login",
    REGISTER: "/api/auth/register",
    LOGOUT: "/api/auth/logout",
    REFRESH: "/api/auth/refresh",
  },
  LOCATIONS: {
    BASE: "/api/locations",
    BY_DEVICE_ID: (deviceId: string) => `/api/locations/${deviceId}`,
  },
  SENSORS: {
    PIS: "/pis",
    PI_DEVICES: (piId: string) => `/pis/${piId}/devices`,
  },
} as const;

export const TOKEN_REFRESH_THRESHOLD = 60 * 1000; // 1 minute before expiry

