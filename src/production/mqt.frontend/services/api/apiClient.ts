import { API_BASE_URL, API_ENDPOINTS, TOKEN_REFRESH_THRESHOLD } from "@/constants/api";
import type { AuthResponse, RefreshTokenResponse } from "@/types/auth";

// Token storage in memory
let accessToken: string | null = null;
let tokenExpiry: number | null = null;

// Token management functions
export function setTokens(authResponse: AuthResponse) {
  accessToken = authResponse.access_token;
  // If expires_at is not provided, default to 1 hour from now
  const expiresAt = authResponse.expires_at || Math.floor(Date.now() / 1000) + 3600;
  tokenExpiry = expiresAt * 1000; // Convert to milliseconds
  // Optionally store in sessionStorage for persistence across page reloads
  if (typeof window !== "undefined") {
    sessionStorage.setItem("access_token", authResponse.access_token);
    sessionStorage.setItem("token_expiry", expiresAt.toString());
    sessionStorage.setItem("user_id", authResponse.user_id);
    sessionStorage.setItem("username", authResponse.username);
    sessionStorage.setItem("email", authResponse.email);
    sessionStorage.setItem("role", authResponse.role);
  }
}

export function getAccessToken(): string | null {
  if (accessToken) return accessToken;
  
  // Try to restore from sessionStorage
  if (typeof window !== "undefined") {
    const stored = sessionStorage.getItem("access_token");
    const expiry = sessionStorage.getItem("token_expiry");
    if (stored && expiry) {
      const expiryTime = parseInt(expiry) * 1000;
      if (expiryTime > Date.now()) {
        accessToken = stored;
        tokenExpiry = expiryTime;
        return stored;
      } else {
        // Token expired, clear storage
        clearTokens();
      }
    }
  }
  return null;
}

export function getTokenExpiry(): number | null {
  return tokenExpiry;
}

export function clearTokens() {
  accessToken = null;
  tokenExpiry = null;
  if (typeof window !== "undefined") {
    sessionStorage.removeItem("access_token");
    sessionStorage.removeItem("token_expiry");
    sessionStorage.removeItem("user_id");
    sessionStorage.removeItem("username");
    sessionStorage.removeItem("email");
    sessionStorage.removeItem("role");
  }
}

export function shouldRefreshToken(): boolean {
  if (!accessToken || !tokenExpiry) return false;
  const now = Date.now();
  const timeUntilExpiry = tokenExpiry - now;
  return timeUntilExpiry < TOKEN_REFRESH_THRESHOLD;
}

// Refresh token function
export async function refreshAccessToken(): Promise<string> {
  try {
    const response = await fetch(`${API_BASE_URL}${API_ENDPOINTS.AUTH.REFRESH}`, {
      method: "POST",
      credentials: "include", // IMPORTANT: Include cookies
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      // If refresh fails with 401, it means there's no valid refresh token
      // This is normal for new logins or expired sessions
      if (response.status === 401) {
        clearTokens();
        throw new Error("No valid refresh token available");
      }
      throw new Error("Token refresh failed");
    }

    const data: RefreshTokenResponse = await response.json();
    
    // Update token in memory
    accessToken = data.access_token;
    tokenExpiry = data.expires_at * 1000;
    
    // Update sessionStorage
    if (typeof window !== "undefined") {
      sessionStorage.setItem("access_token", data.access_token);
      sessionStorage.setItem("token_expiry", data.expires_at.toString());
    }

    return data.access_token;
  } catch (error) {
    console.error("Token refresh failed:", error);
    clearTokens();
    throw error;
  }
}

// Fetch wrapper with automatic token refresh
export async function apiFetch(
  url: string,
  options: RequestInit = {}
): Promise<Response> {
  // Get current token first
  const token = getAccessToken();
  
  // Only try to refresh if we have a token and it needs refreshing
  if (token && shouldRefreshToken()) {
    try {
      await refreshAccessToken();
    } catch (error) {
      // If refresh fails, clear tokens but don't throw - let the request proceed
      // The request will fail with 401 if token is invalid, which we handle below
      clearTokens();
      console.warn("Token refresh failed before request:", error);
    }
  }

  // Get token again (might have been refreshed)
  const currentToken = getAccessToken();

  // Add authorization header if token exists
  const headers = new Headers(options.headers);
  if (currentToken) {
    headers.set("Authorization", `Bearer ${currentToken}`);
  }
  // Only set Content-Type if not already set
  if (!headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  // Make the request
  let response = await fetch(`${API_BASE_URL}${url}`, {
    ...options,
    headers,
    credentials: "include", // Always include cookies
  });

  // If 401, try refreshing token once (only if we had a token to begin with)
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  if (response.status === 401 && !(options as any)._retry && currentToken) {
    try {
      await refreshAccessToken();
      const newToken = getAccessToken();
      
      if (newToken) {
        headers.set("Authorization", `Bearer ${newToken}`);
        // Retry original request with new token
        response = await fetch(`${API_BASE_URL}${url}`, {
          ...options,
          headers,
          credentials: "include",
          _retry: true,
        } as RequestInit);
      }
    } catch (error) {
      // Refresh failed - clear tokens
      clearTokens();
      // For profile/location/weather endpoints, return the 401 response instead of throwing
      // This allows the service layer to handle "no location set" gracefully
      if (url.includes("/profile") || url.includes("/location") || url.includes("/weather") || url.includes("/locations")) {
        return response; // Return original 401 response
      }
      // Don't throw error - let the caller handle the 401 response
      console.warn("Token refresh failed after 401:", error);
    }
  }

  return response;
}

