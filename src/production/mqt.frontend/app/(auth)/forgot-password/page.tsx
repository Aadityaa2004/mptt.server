"use client";

import Link from "next/link";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import Navbar from "@/components/navbar/Navbar";
import { authService } from "@/services/api/authService";

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [submitted, setSubmitted] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setIsLoading(true);
    try {
      await authService.forgotPassword(email);
      setSubmitted(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Request failed. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  if (submitted) {
    return (
      <div className="min-h-screen bg-black text-white">
        <Navbar />
        <main className="flex items-center justify-center min-h-[calc(100vh-4rem)] px-4 py-16">
          <div className="w-full max-w-md text-center space-y-6">
            <h1 className="text-3xl sm:text-4xl font-light tracking-tight">Check your email</h1>
            <p className="text-white/60 font-light text-sm">
              If an account exists with <span className="text-white/80">{email}</span>, you will
              receive a password reset link shortly.
            </p>
            <p className="text-white/60 font-light text-sm">
              The link expires in 1 hour. If you don&apos;t see the email, check your spam folder.
            </p>
            <Link
              href="/login"
              className="inline-block text-white hover:text-white/80 font-light underline underline-offset-2"
            >
              Back to Sign In
            </Link>
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
            <h1 className="text-4xl font-light tracking-tight mb-3">Forgot Password</h1>
            <p className="text-white/60 font-light text-sm">
              Enter your email and we&apos;ll send you a link to reset your password.
            </p>
          </div>

          {error && (
            <div className="mb-4 p-3 bg-red-500/10 border border-red-500/20 rounded-lg">
              <p className="text-sm text-red-400 font-light">{error}</p>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="space-y-2">
              <label htmlFor="email" className="text-sm font-light text-white/80 block">
                Email
              </label>
              <Input
                id="email"
                type="email"
                placeholder="Enter your email"
                value={email}
                onChange={(e) => {
                  setEmail(e.target.value);
                  setError("");
                }}
                required
                className="bg-transparent border-white/20 text-white placeholder:text-white/40 focus:border-white/40 rounded-lg h-11"
              />
            </div>

            <Button
              type="submit"
              disabled={isLoading}
              className="w-full bg-white text-black hover:bg-white/90 rounded-lg h-11 font-light disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? "Sending..." : "Send reset link"}
            </Button>
          </form>

          <div className="relative my-8">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-white/10" />
            </div>
            <div className="relative flex justify-center text-xs">
              <span className="bg-black px-4 text-white/40 font-light">OR</span>
            </div>
          </div>

          <div className="text-center">
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
