"use client";

import Navbar from "@/components/navbar/Navbar";
import { DottedGlowBackground } from "@/components/ui/dotted-glow-background";

export default function ProductsPage() {
  return (
    <div className="relative min-h-screen bg-background text-foreground overflow-hidden">
      <Navbar />

      <main className="relative z-20 px-4 sm:px-6 lg:px-8 pt-16 sm:pt-24 pb-12 sm:pb-20">
        <div className="container mx-auto max-w-5xl">
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
                MapleSense Products
              </h1>
              <p className="text-white/70 font-light text-sm max-w-2xl mx-auto">
                A preview of the hardware and software components that make up
                the MapleSense monitoring system.
              </p>
            </div>
          </div>

          <section className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="p-6 rounded-lg border border-white/10 bg-black/40 backdrop-blur-sm">
              <h2 className="text-lg font-light mb-2">Bucket Sensors</h2>
              <p className="text-sm text-white/70 font-light">
                Weatherproof ultrasonic sensors mounted on traditional buckets
                to monitor sap levels, temperature, and battery status.
              </p>
            </div>

            <div className="p-6 rounded-lg border border-white/10 bg-black/40 backdrop-blur-sm">
              <h2 className="text-lg font-light mb-2">Raspberry Pi Gateway</h2>
              <p className="text-sm text-white/70 font-light">
                A central controller that aggregates wireless sensor data and
                securely forwards it to the cloud via MQTT and our REST API.
              </p>
            </div>

            <div className="p-6 rounded-lg border border-white/10 bg-black/40 backdrop-blur-sm">
              <h2 className="text-lg font-light mb-2">Web Dashboard</h2>
              <p className="text-sm text-white/70 font-light">
                A modern, responsive dashboard that visualizes your operation,
                including interactive maps, trends, and recommendations.
              </p>
            </div>
          </section>
        </div>
      </main>
    </div>
  );
}

