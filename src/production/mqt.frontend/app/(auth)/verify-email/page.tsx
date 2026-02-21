"use client";

import Link from "next/link";
import { useState, useEffect, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import Navbar from "@/components/navbar/Navbar";
import { useAuth } from "@/contexts/AuthContext";
import { authService } from "@/services/api/authService";

function VerifyEmailContent() {
  const [otp, setOtp] = useState("");
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [resendCooldown, setResendCooldown] = useState(0);
  const { verifyEmail, user, isAuthenticated, isLoading: authLoading } = useAuth();
  const router = useRouter();
  const searchParams = useSearchParams();
  const email = searchParams.get("email") || "";

  useEffect(() => {
    if (!authLoading && isAuthenticated && user) {
      if (user.role === "admin") {
        router.push("/admin/dashboard");
      } else {
        router.push("/user/dashboard");
      }
    }
  }, [isAuthenticated, user, authLoading, router]);

  useEffect(() => {
    if (resendCooldown > 0) {
      const t = setTimeout(() => setResendCooldown(resendCooldown - 1), 1000);
      return () => clearTimeout(t);
    }
  }, [resendCooldown]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    if (!email) {
      setError("Email is required. Please go back to registration.");
      return;
    }
    if (otp.length !== 6) {
      setError("Please enter the 6-digit code from your email.");
      return;
    }
    setIsLoading(true);
    try {
      await verifyEmail(email, otp);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Verification failed. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  const handleResend = async () => {
    if (resendCooldown > 0 || !email) return;
    setError("");
    try {
      await authService.resendOtp(email);
      setResendCooldown(60);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to resend code.");
    }
  };

  if (authLoading) {
    return (
      <div className="min-h-screen bg-background text-foreground flex items-center justify-center">
        <p className="text-white/60 font-light">Loading...</p>
      </div>
    );
  }

  if (isAuthenticated) {
    return null;
  }

  if (!email) {
    return (
      <div className="min-h-screen bg-background text-foreground">
        <Navbar />
        <main className="flex items-center justify-center min-h-[calc(100vh-4rem)] px-4 py-12 sm:py-16">
          <div className="text-center">
            <h1 className="text-2xl font-light mb-4">Invalid verification link</h1>
            <p className="text-white/60 mb-6">
              Please complete registration first to receive a verification code.
            </p>
            <Link href="/register" className="text-white underline">
              Go to registration
            </Link>
          </div>
        </main>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background text-foreground">
      <Navbar />
      <main className="flex items-center justify-center min-h-[calc(100vh-4rem)] px-4 py-12 sm:py-16">
        <div className="w-full max-w-md">
          <div className="text-center mb-12">
            <h1 className="text-2xl sm:text-4xl font-light tracking-tight mb-3">Verify your email</h1>
            <p className="text-white/60 font-light text-sm">
              We sent a 6-digit code to <span className="text-white/80">{email}</span>
            </p>
          </div>

          {error && (
            <div className="mb-4 p-3 bg-red-500/10 border border-red-500/20 rounded-lg">
              <p className="text-sm text-red-400 font-light">{error}</p>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="space-y-2">
              <label htmlFor="otp" className="text-sm font-light text-white/80 block">
                Verification code
              </label>
              <Input
                id="otp"
                type="text"
                inputMode="numeric"
                maxLength={6}
                placeholder="000000"
                value={otp}
                onChange={(e) => {
                  const v = e.target.value.replace(/\D/g, "");
                  setOtp(v);
                  setError("");
                }}
                className="bg-transparent border-white/20 text-white placeholder:text-white/40 focus:border-white/40 rounded-lg h-11 text-center text-2xl tracking-[0.5em] font-light"
              />
            </div>

            <Button
              type="submit"
              disabled={isLoading || otp.length !== 6}
              className="w-full bg-white text-black hover:bg-white/90 rounded-lg h-11 font-light disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? "Verifying..." : "Verify"}
            </Button>
          </form>

          <div className="mt-6 text-center">
            <p className="text-sm text-white/60 font-light">
              Didn&apos;t receive the code?{" "}
              <button
                type="button"
                onClick={handleResend}
                disabled={resendCooldown > 0}
                className="text-white hover:text-white/80 underline disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {resendCooldown > 0 ? `Resend in ${resendCooldown}s` : "Resend"}
              </button>
            </p>
          </div>

          <div className="relative my-8">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-white/10" />
            </div>
            <div className="relative flex justify-center text-xs">
              <span className="bg-black px-4 text-white/40 font-light">OR</span>
            </div>
          </div>

          <div className="text-center">
            <Link href="/login" className="text-sm text-white/60 hover:text-white font-light underline">
              Back to Sign In
            </Link>
          </div>
        </div>
      </main>
    </div>
  );
}

export default function VerifyEmailPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen bg-background text-foreground flex items-center justify-center">
          <p className="text-white/60 font-light">Loading...</p>
        </div>
      }
    >
      <VerifyEmailContent />
    </Suspense>
  );
}
