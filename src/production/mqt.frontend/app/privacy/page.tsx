"use client";

import Navbar from "@/components/navbar/Navbar";

export default function PrivacyPolicyPage() {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <Navbar />

      <main className="px-4 sm:px-6 lg:px-8 pt-16 sm:pt-24 pb-12 sm:pb-20">
        <div className="container mx-auto max-w-3xl space-y-8">
          <header>
            <h1 className="text-3xl sm:text-4xl font-light mb-3">Privacy Policy</h1>
            <p className="text-sm text-white/60 font-light">
              This is a placeholder privacy statement for the MapleSense project.
              It can be expanded with formal legal language as needed.
            </p>
          </header>

          <section className="space-y-4 text-sm text-white/70 font-light">
            <p>
              MapleSense collects sensor data related to sap levels, temperature,
              and device status for the purpose of providing monitoring and
              analytics to maple producers. Basic account information such as
              username and email address is used to authenticate users and
              deliver the service.
            </p>
            <p>
              This deployment is currently an academic capstone project. Any
              production use should be accompanied by an updated privacy policy
              that reflects real data handling practices, retention, and user
              rights.
            </p>
          </section>
        </div>
      </main>
    </div>
  );
}

