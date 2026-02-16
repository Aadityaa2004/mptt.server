import { API_ENDPOINTS } from "@/constants/api";
import type { LoginRequest, RegisterRequest, AuthResponse } from "@/types/auth";
import { apiFetch, setTokens, clearTokens } from "./apiClient";

export const authService = {
  async login(credentials: LoginRequest): Promise<AuthResponse> {
    const response = await apiFetch(API_ENDPOINTS.AUTH.LOGIN, {
      method: "POST",
      body: JSON.stringify(credentials),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: "Login failed" }));
      throw new Error(error.message || "Login failed");
    }

    const data: AuthResponse = await response.json();
    setTokens(data);
    return data;
  },

  async register(userData: RegisterRequest): Promise<AuthResponse | null> {
    // Ensure role is set to "user" by default
    const payload = {
      ...userData,
      role: "user" as const,
    };

    const response = await apiFetch(API_ENDPOINTS.AUTH.REGISTER, {
      method: "POST",
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: "Registration failed" }));
      throw new Error(error.message || "Registration failed");
    }

    const data: AuthResponse = await response.json();
    // Only set tokens if access_token is provided (backend may not provide tokens on registration)
    if (data.access_token) {
      setTokens(data);
      return data;
    }
    // Return null if no tokens provided (user needs to login)
    return null;
  },

  async logout(): Promise<void> {
    try {
      await apiFetch(API_ENDPOINTS.AUTH.LOGOUT, {
        method: "POST",
      });
    } catch (error) {
      console.error("Logout error:", error);
    } finally {
      clearTokens();
    }
  },
};

