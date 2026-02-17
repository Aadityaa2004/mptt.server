"use client";

import Navbar from "@/components/navbar/Navbar";

export default function TermsOfUsePage() {
  return (
    <div className="min-h-screen bg-black text-white">
      <Navbar />

      <main className="px-4 sm:px-6 lg:px-8 pt-24 pb-20">
        <div className="container mx-auto max-w-3xl space-y-8">
          <header>
            <h1 className="text-3xl sm:text-4xl font-light mb-3">Terms of Use</h1>
            <p className="text-sm text-white/60 font-light">
              These terms describe the intended use of the MapleSense web
              application in its current project state.
            </p>
          </header>

          <section className="space-y-4 text-sm text-white/70 font-light">
            <p>
              The MapleSense system is provided as part of an academic capstone
              project and is intended for demonstration and evaluation purposes
              only. No guarantees are made regarding uptime, data retention, or
              suitability for production use.
            </p>
            <p>
              By using this site, you acknowledge that sensor readings and
              analytics are provided on a best-effort basis and should not be
              relied upon as the sole source of operational decision-making
              without additional verification.
            </p>
          </section>
        </div>
      </main>
    </div>
  );
}

