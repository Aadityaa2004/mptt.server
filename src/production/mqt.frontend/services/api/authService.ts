import { API_ENDPOINTS } from "@/constants/api";
import type {
  LoginRequest,
  RegisterRequest,
  AuthResponse,
  RegisterResponse,
} from "@/types/auth";
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

  async register(userData: RegisterRequest): Promise<RegisterResponse> {
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
      throw new Error(error.error || error.message || "Registration failed");
    }

    return (await response.json()) as RegisterResponse;
  },

  async verifyEmail(email: string, otp: string): Promise<AuthResponse> {
    const response = await apiFetch(API_ENDPOINTS.AUTH.VERIFY_EMAIL, {
      method: "POST",
      body: JSON.stringify({ email, otp }),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: "Verification failed" }));
      throw new Error(error.error || error.message || "Verification failed");
    }

    const data: AuthResponse = await response.json();
    setTokens(data);
    return data;
  },

  async resendOtp(email: string): Promise<void> {
    const response = await apiFetch(API_ENDPOINTS.AUTH.RESEND_OTP, {
      method: "POST",
      body: JSON.stringify({ email }),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: "Failed to resend OTP" }));
      throw new Error(error.error || error.message || "Failed to resend OTP");
    }
  },

  async forgotPassword(email: string): Promise<void> {
    const response = await apiFetch(API_ENDPOINTS.AUTH.FORGOT_PASSWORD, {
      method: "POST",
      body: JSON.stringify({ email }),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: "Request failed" }));
      throw new Error(error.error || error.message || "Request failed");
    }
  },

  async resetPassword(token: string, newPassword: string): Promise<void> {
    const response = await apiFetch(API_ENDPOINTS.AUTH.RESET_PASSWORD, {
      method: "POST",
      body: JSON.stringify({ token, new_password: newPassword }),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: "Password reset failed" }));
      throw new Error(error.error || error.message || "Password reset failed");
    }
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

