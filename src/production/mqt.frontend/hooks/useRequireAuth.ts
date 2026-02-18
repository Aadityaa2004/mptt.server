"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/contexts/AuthContext";

export function useRequireAuth(requiredRole?: "user" | "admin") {
  const { user, isLoading, isAuthenticated } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading) {
      if (!isAuthenticated) {
        router.push("/login");
        return;
      }

      if (requiredRole && user?.role !== requiredRole) {
        // Redirect to appropriate dashboard based on role
        if (user?.role === "admin") {
          router.push("/admin/dashboard");
        } else {
          router.push("/user/dashboard");
        }
      }
    }
  }, [isLoading, isAuthenticated, user, requiredRole, router]);

  return { user, isLoading, isAuthenticated };
}

