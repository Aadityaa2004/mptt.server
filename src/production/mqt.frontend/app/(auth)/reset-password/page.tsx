"use client";

import Link from "next/link";
import { useState, useEffect, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import Navbar from "@/components/navbar/Navbar";
import { authService } from "@/services/api/authService";

function ResetPasswordContent() {
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get("token") || "";

  useEffect(() => {
    if (!token) {
      setError("Invalid or missing reset link. Please request a new password reset.");
    }
  }, [token]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    if (password.length < 6) {
      setError("Password must be at least 6 characters");
      return;
    }
    if (password !== confirmPassword) {
      setError("Passwords do not match");
      return;
    }
    if (!token) return;
    setIsLoading(true);
    try {
      await authService.resetPassword(token, password);
      setSuccess(true);
      setTimeout(() => router.push("/login"), 2000);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Password reset failed. The link may have expired.");
    } finally {
      setIsLoading(false);
    }
  };

  if (success) {
    return (
      <div className="min-h-screen bg-black text-white">
        <Navbar />
        <main className="flex items-center justify-center min-h-[calc(100vh-4rem)] px-4 py-16">
          <div className="text-center space-y-6">
            <h1 className="text-3xl font-light tracking-tight">Password reset successful</h1>
            <p className="text-white/60 font-light">Redirecting you to sign in...</p>
          </div>
        </main>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-black text-white">
      <Navbar />
      <main className="flex items-center justify-center min-h-[calc(100vh-4rem)] px-4 py-16">
        <div className="w-full max-w-md">
          <div className="text-center mb-12">
            <h1 className="text-4xl font-light tracking-tight mb-3">Set new password</h1>
            <p className="text-white/60 font-light text-sm">
              Enter your new password below.
            </p>
          </div>

          {error && (
            <div className="mb-4 p-3 bg-red-500/10 border border-red-500/20 rounded-lg">
              <p className="text-sm text-red-400 font-light">{error}</p>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="space-y-2">
              <label htmlFor="password" className="text-sm font-light text-white/80 block">
                New password
              </label>
              <Input
                id="password"
                type="password"
                placeholder="Enter new password"
                value={password}
                onChange={(e) => {
                  setPassword(e.target.value);
                  setError("");
                }}
                required
                minLength={6}
                className="bg-transparent border-white/20 text-white placeholder:text-white/40 focus:border-white/40 rounded-lg h-11"
              />
            </div>

            <div className="space-y-2">
              <label htmlFor="confirmPassword" className="text-sm font-light text-white/80 block">
                Confirm password
              </label>
              <Input
                id="confirmPassword"
                type="password"
                placeholder="Confirm new password"
                value={confirmPassword}
                onChange={(e) => {
                  setConfirmPassword(e.target.value);
                  setError("");
                }}
                required
                minLength={6}
                className="bg-transparent border-white/20 text-white placeholder:text-white/40 focus:border-white/40 rounded-lg h-11"
              />
            </div>

            <Button
              type="submit"
              disabled={isLoading || !token}
              className="w-full bg-white text-black hover:bg-white/90 rounded-lg h-11 font-light disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? "Resetting..." : "Reset password"}
            </Button>
          </form>

          <div className="mt-8 text-center">
            <Link
              href="/login"
              className="text-sm text-white/60 hover:text-white font-light underline underline-offset-2"
            >
              Back to Sign In
            </Link>
          </div>
        </div>
      </main>
    </div>
  );
}

export default function ResetPasswordPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen bg-black text-white flex items-center justify-center">
          <p className="text-white/60 font-light">Loading...</p>
        </div>
      }
    >
      <ResetPasswordContent />
    </Suspense>
  );
}
