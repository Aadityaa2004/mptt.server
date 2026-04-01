// Base URLs should NOT include the `/api` prefix, because all
// endpoint paths below already start with `/api/...`.
// Local: use http://localhost:9002 (direct to API) or http://localhost (via nginx).
// Production: set NEXT_PUBLIC_* at build time (e.g. https://orpheus-networks.com).

/** Strips a trailing `/api` so misconfigured build-args or env still produce correct URLs. */
function normalizeApiPublicBaseUrl(raw: string): string {
  const trimmed = raw.trim().replace(/\/+$/, "");
  if (trimmed.endsWith("/api")) {
    const withoutApi = trimmed.slice(0, -4).replace(/\/+$/, "");
    return withoutApi || trimmed;
  }
  return trimmed;
}

export const API_BASE_URL = normalizeApiPublicBaseUrl(
  process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:9002"
);

export const READINGS_API_BASE_URL = normalizeApiPublicBaseUrl(
  process.env.NEXT_PUBLIC_READINGS_API_BASE_URL || "http://localhost:9002"
);

export const API_ENDPOINTS = {
  AUTH: {
    LOGIN: "/api/auth/login",
    REGISTER: "/api/auth/register",
    LOGOUT: "/api/auth/logout",
    REFRESH: "/api/auth/refresh",
    VERIFY_EMAIL: "/api/auth/verify-email",
    RESEND_OTP: "/api/auth/resend-otp",
    FORGOT_PASSWORD: "/api/auth/forgot-password",
    RESET_PASSWORD: "/api/auth/reset-password",
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

