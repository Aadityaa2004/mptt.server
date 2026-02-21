"use client";

import Navbar from "@/components/navbar/Navbar";
import { DottedGlowBackground } from "@/components/ui/dotted-glow-background";

export default function ContactUsPage() {
  return (
    <div className="relative min-h-screen bg-background text-foreground overflow-hidden">
      <Navbar />

      <main className="relative z-20 px-4 sm:px-6 lg:px-8 pt-16 sm:pt-24 pb-12 sm:pb-20">
        <div className="container mx-auto max-w-4xl">
          <div className="relative mb-16 pt-12 pb-12 rounded-2xl overflow-hidden">
            <div className="absolute inset-0 overflow-hidden pointer-events-none z-0">
              <DottedGlowBackground
                className="pointer-events-none mask-radial-to-90% mask-radial-at-center opacity-20 dark:opacity-100"
                opacity={1}
                gap={10}
                radius={1.6}
                colorLightVar="--color-neutral-500"
                glowColorLightVar="--color-neutral-600"
                colorDarkVar="--color-neutral-500"
                glowColorDarkVar="--color-sky-800"
                backgroundOpacity={0}
                speedMin={0.3}
                speedMax={1.6}
                speedScale={1}
              />
            </div>

            <div className="relative z-10 text-center space-y-4">
              <h1 className="text-4xl sm:text-5xl font-light tracking-tight">
                Contact Us
              </h1>
              <p className="text-white/70 font-light text-sm max-w-xl mx-auto">
                Have questions about MapleSense, deployment, or pilot programs?
                Reach out and our team will get back to you.
              </p>
            </div>
          </div>

          <section className="space-y-6">
            <div className="p-6 rounded-lg border border-white/10 bg-black/40 backdrop-blur-sm">
              <h2 className="text-xl font-light mb-3">Email</h2>
              <p className="text-sm text-white/70 font-light">
                For general questions and support, email us at{" "}
                <span className="underline underline-offset-2">
                  maplesense2025@gmail.com
                </span>
                .
              </p>
            </div>

            <div className="p-6 rounded-lg border border-white/10 bg-black/40 backdrop-blur-sm">
              <h2 className="text-xl font-light mb-3">Project Team</h2>
              <p className="text-sm text-white/70 font-light">
                MapleSense is an academic capstone project developed by a team
                of Computer Engineering students. This page will be updated with
                additional contact options as the project evolves.
              </p>
            </div>
          </section>
        </div>
      </main>
    </div>
  );
}

