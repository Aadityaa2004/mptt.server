"use client";

import Navbar from "@/components/navbar/Navbar";

export default function ForgotPasswordPage() {
  return (
    <div className="min-h-screen bg-black text-white">
      <Navbar />

      <main className="flex items-center justify-center min-h-[calc(100vh-4rem)] px-4 py-16">
        <div className="w-full max-w-md text-center space-y-6">
          <h1 className="text-3xl sm:text-4xl font-light tracking-tight">
            Forgot Password
          </h1>
          <p className="text-white/60 font-light text-sm">
            Password reset is not yet implemented in this deployment. If you
            need help accessing your account, please contact the MapleSense team
            using the email on the Contact Us page.
          </p>
        </div>
      </main>
    </div>
  );
}

