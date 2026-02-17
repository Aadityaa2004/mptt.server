"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import Navbar from "@/components/navbar/Navbar";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/contexts/AuthContext";

export default function MapLandingPage() {
  const { user, isAuthenticated, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && isAuthenticated && user) {
      if (user.role === "admin") {
        router.replace("/admin/dashboard");
      } else {
        router.replace("/user/dashboard");
      }
    }
  }, [isAuthenticated, isLoading, user, router]);

  return (
    <div className="min-h-screen bg-black text-white">
      <Navbar />

      <main className="flex items-center justify-center min-h-[calc(100vh-4rem)] px-4 py-16">
        <div className="w-full max-w-xl text-center space-y-6">
          <h1 className="text-4xl font-light tracking-tight">
            View Your MapleSense Dashboard
          </h1>
          <p className="text-white/60 font-light text-sm max-w-md mx-auto">
            To access the interactive map and real-time sensor dashboard, please
            sign in to your MapleSense account. Once authenticated, you&apos;ll be
            redirected automatically based on your role.
          </p>

          <div className="flex flex-col sm:flex-row items-center justify-center gap-4 mt-8">
            <Link href="/login">
              <Button className="bg-white text-black hover:bg-white/90 rounded-lg h-11 px-8 font-light">
                Login to Dashboard
              </Button>
            </Link>
            <Link href="/register">
              <Button
                variant="outline"
                className="text-white border-2 border-white hover:bg-white/10 rounded-lg h-11 px-8 font-light"
              >
                Create Account
              </Button>
            </Link>
          </div>
        </div>
      </main>
    </div>
  );
}

