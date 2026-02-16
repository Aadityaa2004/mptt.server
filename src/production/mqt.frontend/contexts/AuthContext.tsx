"use client";

import React, { createContext, useContext, useState, useEffect, useCallback, useRef } from "react";
import { useRouter } from "next/navigation";
import type { User, UserRole, AuthResponse } from "@/types/auth";
import { authService } from "@/services/api/authService";
import { getAccessToken, shouldRefreshToken, refreshAccessToken, clearTokens } from "@/services/api/apiClient";

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (username: string, password: string) => Promise<void>;
  register: (username: string, email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  checkAuth: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const router = useRouter();
  const refreshIntervalRef = useRef<NodeJS.Timeout | null>(null);

  const checkAuth = useCallback(() => {
    if (typeof window === "undefined") {
      setIsLoading(false);
      return;
    }

    const token = getAccessToken();
    const userId = sessionStorage.getItem("user_id");
    const username = sessionStorage.getItem("username");
    const email = sessionStorage.getItem("email");
    const role = sessionStorage.getItem("role") as UserRole | null;

    if (token && userId && username && email && role) {
      setUser({
        user_id: userId,
        username,
        email,
        role,
      });
    } else {
      setUser(null);
    }
    setIsLoading(false);
  }, []);

  const handleLogout = useCallback(async () => {
    // Clear the token refresh interval
    if (refreshIntervalRef.current) {
      clearInterval(refreshIntervalRef.current);
      refreshIntervalRef.current = null;
    }
    
    // Clear tokens from memory and storage
    clearTokens();
    
    await authService.logout();
    setUser(null);
    router.push("/login");
  }, [router]);

  // Initialize auth state from sessionStorage
  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    checkAuth();
  }, [checkAuth]);

  // Set up token refresh interval when user is authenticated
  useEffect(() => {
    // Clear any existing interval first
    if (refreshIntervalRef.current) {
      clearInterval(refreshIntervalRef.current);
      refreshIntervalRef.current = null;
    }

    // Only set up interval if user is authenticated
    if (user) {
      refreshIntervalRef.current = setInterval(() => {
        if (shouldRefreshToken()) {
          refreshAccessToken().catch(() => {
            // If refresh fails, logout
            handleLogout();
          });
        }
      }, 60000); // Check every minute
    }

    return () => {
      if (refreshIntervalRef.current) {
        clearInterval(refreshIntervalRef.current);
        refreshIntervalRef.current = null;
      }
    };
  }, [user, handleLogout]);

  const login = useCallback(async (username: string, password: string) => {
    try {
      // Clear any existing tokens and interval before logging in
      if (refreshIntervalRef.current) {
        clearInterval(refreshIntervalRef.current);
        refreshIntervalRef.current = null;
      }
      clearTokens();
      
      const response: AuthResponse = await authService.login({ username, password });
      
      setUser({
        user_id: response.user_id,
        username: response.username,
        email: response.email,
        role: response.role,
      });

      // Navigate based on role
      if (response.role === "admin") {
        router.push("/admin/dashboard");
      } else {
        router.push("/user/dashboard");
      }
    } catch (error) {
      throw error;
    }
  }, [router]);

  const register = useCallback(async (username: string, email: string, password: string) => {
    try {
      // Clear any existing tokens and interval before registering
      if (refreshIntervalRef.current) {
        clearInterval(refreshIntervalRef.current);
        refreshIntervalRef.current = null;
      }
      clearTokens();
      
      const response = await authService.register({
        username,
        email,
        password,
      });

      // If registration doesn't provide tokens (backend design), redirect to login
      if (!response || !response.access_token) {
        // Don't set user state - they need to login first
        router.push("/login");
        return;
      }

      setUser({
        user_id: response.user_id,
        username: response.username,
        email: response.email,
        role: response.role,
      });

      // Navigate based on role (should always be "user" for registration)
      if (response.role === "admin") {
        router.push("/admin/dashboard");
      } else {
        router.push("/user/dashboard");
      }
    } catch (error) {
      throw error;
    }
  }, [router]);

  const logout = useCallback(async () => {
    await handleLogout();
  }, [handleLogout]);

  return (
    <AuthContext.Provider
      value={{
        user,
        isLoading,
        isAuthenticated: !!user,
        login,
        register,
        logout,
        checkAuth,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}

