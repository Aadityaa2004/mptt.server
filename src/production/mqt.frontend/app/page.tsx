"use client";

import Link from "next/link";
import { ExternalLink, ArrowRight, ChevronDown } from "lucide-react";
import { Button } from "@/components/ui/button";
import Navbar from "@/components/navbar/Navbar";
import { DottedGlowBackground } from "@/components/ui/dotted-glow-background";

export default function Home() {
  return (
    <div className="relative min-h-screen bg-black text-white overflow-hidden">
      <Navbar />

      {/* Main Content */}
      <main className="relative z-20  px-4 sm:px-6 lg:px-8">
        <div className="container mx-auto max-w-6xl">
          {/* Hero Section with DottedGlowBackground */}
          <div className="relative mb-1 pt-32 pb-1 min-h-[55vh] rounded-2xl overflow-hidden">
            {/* Background Effects - Only for Hero Section */}
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

            {/* Hero Content */}
            <div className="relative z-10 h-full flex flex-col justify-start">
              {/* Monitor Sap Collection Button */}
              <div className="flex justify-center mb-8">
                {/* <Button
                  variant="ghost"
                  className="rounded-full border border-white/20 hover:border-white/40 hover:bg-white/5 text-white/90 text-sm font-light px-6 h-10 "
                >
                  <div className="w-4 h-4 rounded-full border border-white/40 mr-2 flex items-center justify-center">
                    <ArrowRight className="h-2.5 w-2.5" />
                  </div>
                  Monitor Your Sap Collection!
                  <ArrowRight className="h-3.5 w-3.5 ml-2" />
                </Button> */}
              </div>

              {/* Main Headline */}
              <div className="text-center mb-6">
                <h1 className="text-4xl sm:text-5xl lg:text-6xl font-light tracking-tight mb-4">
                  Introducing MapleSense
                </h1>
                <p className="text-lg sm:text-xl text-white/80 font-light max-w-2xl mx-auto">
                  Optimize and simplify maple sap collection for small-scale and hobbyist producers using bucket and tap setups. Monitor sap levels remotely, track weather conditions, and receive intelligent collection recommendations—all from your web dashboard.
                </p>
              </div>

              {/* CTA Buttons */}
              <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
                <Link href="/map">
                  <Button
                    variant="outline"
                    className="rounded-lg border-white/20 backdrop-blur-sm hover:text-white/55 text-white font-light px-8 h-10 w-full sm:w-auto"
                  >
                    View Dashboard
                    <ExternalLink className="h-3 w-4 ml-2" />
                  </Button>
                </Link>
                <Button
                  variant="default"
                  className="rounded-lg bg-orange-500/85 text-white hover:bg-orange-500/70 font-light px-8 h-10 w-full sm:w-auto"
                >
                  Learn More
                </Button>
              </div>
            </div>
          </div>

          {/* Video Section */}
          <div className="relative mb-16">
            <div className="relative w-full aspect-video rounded-lg overflow-hidden border border-white/10 bg-black/40 backdrop-blur-sm">
              <video
                className="w-full h-full object-cover"
                src="/sample.mp4"
                autoPlay
                loop
                muted
                playsInline
              />
            </div>
          </div>

          {/* Key Features */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-20 max-w-6xl mx-auto relative z-30">
            <DataPoint 
              name="Real-Time Monitoring" 
              value="24/7 Access" 
              description="Monitor sap levels, temperature, and battery health remotely through our web dashboard"
            />
            <DataPoint 
              name="Weather Intelligence" 
              value="Smart Predictions" 

              description="AI-powered insights predict optimal collection windows based on weather patterns and environmental conditions"
            />
            <DataPoint 
              name="Ultra Low Power" 
              value="3 Month Battery" 

              description="ESP32-C6 sensors operate efficiently, lasting the entire maple tap season on a single charge"
            />
            <DataPoint 
              name="Scalable Network" 
              value="Wi-Fi 6 Enabled" 

              description="Monitor unlimited buckets with one Raspberry Pi controller using our MQTT-based wireless network"
            />
          </div>

          {/* Problem Statement Section */}
          <section className="my-48 max-w-4xl mx-auto relative z-30">
            <div className="text-center mb-12">
              <h2 className="text-3xl sm:text-4xl font-light mb-4">The Challenge</h2>
              <p className="text-lg text-white/70 font-light max-w-2xl mx-auto">
                Traditional maple sap collection requires manually checking each bucket multiple times a day—a physically demanding and time-consuming process that becomes increasingly inefficient as your tapping area expands.
              </p>
            </div>
            <div className="grid md:grid-cols-2 gap-8">
              <div className="p-6 rounded-lg border border-white/10 bg-black/40 backdrop-blur-sm">
                <h3 className="text-xl font-light mb-3">For Small-Scale Producers</h3>
                <p className="text-white/60 font-light leading-relaxed">
                  While commercial operations use expensive vacuum tubing systems, small-scale and hobbyist producers rely on traditional bucket setups. Existing monitoring solutions cost thousands of dollars and aren't designed for bucket-based collection.
                </p>
              </div>
              <div className="p-6 rounded-lg border border-white/10 bg-black/40 backdrop-blur-sm">
                <h3 className="text-xl font-light mb-3">Harsh Winter Conditions</h3>
                <p className="text-white/60 font-light leading-relaxed">
                  Sap collection occurs during harsh winter months, making frequent physical monitoring difficult. Manual checks risk missing overflow, fallen buckets, or sap spoiling—leading to lost yield and wasted effort.
                </p>
              </div>
            </div>
          </section>

          {/* How It Works Section */}
          <section className="mb-20 mt-48 max-w-6xl mx-auto relative z-30">
            <div className="text-center mb-12">
              <h2 className="text-3xl sm:text-4xl font-light mb-4">How MapleSense Works</h2>
              <p className="text-lg text-white/70 font-light max-w-2xl mx-auto">
                A complete monitoring solution designed specifically for bucket and tap setups
              </p>
            </div>

            {/* Architecture Diagram */}
            <div className="mb-48 p-6 sm:p-8 rounded-lg border border-white/10 bg-black/40 backdrop-blur-sm ">
              <div className="text-center mb-8">
                <h3 className="text-xl font-light mb-2">How Your Data Flows</h3>
                <p className="text-sm text-white/50">From your buckets to your dashboard</p>
              </div>

              {/* Simplified Flow */}
              <div className="flex flex-col md:flex-row items-center justify-center gap-6 md:gap-8">
                {/* Sensors */}
                <div className="flex-1 max-w-[280px] p-5 rounded-lg border border-white/20 bg-black/20 ">
                  <div className="text-center mb-4">
                    <h4 className="text-base font-light text-white/80 mb-2">Your Buckets</h4>
                  </div>
                  <div className="p-4 rounded border border-white/10 bg-white/5">
                    <div className="text-sm font-semibold text-white/80 mb-3">Sensors</div>
                    <div className="text-xs text-white/50 space-y-2">
                      <div className="flex items-center gap-2">
                        <div className="w-1.5 h-1.5 rounded-full bg-white/40"></div>
                        <span>Measure sap levels</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <div className="w-1.5 h-1.5 rounded-full bg-white/40"></div>
                        <span>Track temperature</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <div className="w-1.5 h-1.5 rounded-full bg-white/40"></div>
                        <span>Monitor battery health</span>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Arrow */}
                <div className="flex flex-col items-center justify-center">
                  <ArrowRight className="h-8 w-8 text-white/50 rotate-90 md:rotate-0" />
                  <span className="text-xs text-white/40 mt-2">Wireless</span>
                </div>

                {/* Controller */}
                <div className="flex-1 max-w-[280px] p-5 rounded-lg border border-white/20 bg-black/20 ">
                  <div className="text-center mb-4">
                    <h4 className="text-base font-light text-white/80 mb-2">Controller</h4>
                  </div>
                  <div className="p-4 rounded border border-white/10 bg-white/5">
                    <div className="text-sm font-semibold text-white/80 mb-3">Raspberry Pi</div>
                    <div className="text-xs text-white/50 space-y-2">
                      <div className="flex items-center gap-2">
                        <div className="w-1.5 h-1.5 rounded-full bg-white/40"></div>
                        <span>Collects sensor data</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <div className="w-1.5 h-1.5 rounded-full bg-white/40"></div>
                        <span>Sends to cloud</span>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Arrow */}
                <div className="flex flex-col items-center justify-center">
                  <ArrowRight className="h-8 w-8 text-white/50 rotate-90 md:rotate-0" />
                  <span className="text-xs text-white/40 mt-2">Secure</span>
                </div>

                {/* Cloud */}
                <div className="flex-1 max-w-[280px] p-5 rounded-lg border border-white/20 bg-black/20 ">
                  <div className="text-center mb-4">
                    <h4 className="text-base font-light text-white/80 mb-2">Cloud Processing</h4>
                  </div>
                  <div className="p-4 rounded border border-white/10 bg-white/5">
                    <div className="text-sm font-semibold text-white/80 mb-3">Analytics</div>
                    <div className="text-xs text-white/50 space-y-2">
                      <div className="flex items-center gap-2">
                        <div className="w-1.5 h-1.5 rounded-full bg-white/40"></div>
                        <span>Processes data</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <div className="w-1.5 h-1.5 rounded-full bg-white/40"></div>
                        <span>Weather predictions</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <div className="w-1.5 h-1.5 rounded-full bg-white/40"></div>
                        <span>Smart alerts</span>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Arrow */}
                <div className="flex flex-col items-center justify-center">
                  <ArrowRight className="h-8 w-8 text-white/50 rotate-90 md:rotate-0" />
                  <span className="text-xs text-white/40 mt-2">Real-time</span>
                </div>

                {/* Dashboard */}
                <div className="flex-1 max-w-[280px] p-5 rounded-lg border border-white/20 bg-black/20 ">
                  <div className="text-center mb-4">
                    <h4 className="text-base font-light text-white/80 mb-2">Your Dashboard</h4>
                  </div>
                  <div className="p-4 rounded border border-white/10 bg-white/5">
                    <div className="text-sm font-semibold text-white/80 mb-3">Web Access</div>
                    <div className="text-xs text-white/50 space-y-2">
                      <div className="flex items-center gap-2">
                        <div className="w-1.5 h-1.5 rounded-full bg-white/40"></div>
                        <span>View all sensors</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <div className="w-1.5 h-1.5 rounded-full bg-white/40"></div>
                        <span>Collection insights</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <div className="w-1.5 h-1.5 rounded-full bg-white/40"></div>
                        <span>Anywhere, anytime</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* Data Flow Labels */}
              <div className="mt-8 pt-6 border-t border-white/10">
                <div className="flex flex-wrap justify-center gap-6 text-xs text-white/50">
                  <div className="flex items-center gap-2">
                    <div className="w-2 h-2 rounded-full bg-orange-400"></div>
                    <span className="text-white/80">Sensor data collection</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <ArrowRight className="h-3 w-3" />
                    <span className="text-white/80">Wireless transmission</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="w-2 h-2 rounded-full bg-orange-400"></div>
                    <span className="text-white/80">Cloud processing & analytics</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <ArrowRight className="h-3 w-3" />
                    <span className="text-white/80">Real-time dashboard access</span>
                  </div>
                </div>
              </div>
            </div>

            <div className="space-y-8">
              {/* Step 01 - Video on right, text on left */}
              <div className="p-6 rounded-lg border border-white/10 bg-black/40 backdrop-blur-sm">
                <div className="grid md:grid-cols-2 gap-6 items-center">
                  <div className="flex items-start gap-4">
                    <div className="text-2xl font-light text-white/40">01</div>
                    <div>
                      <h3 className="text-xl font-light mb-2">Sensor Deployment</h3>
                      <p className="text-white/60 font-light leading-relaxed">
                        Weatherproof ESP32-C6 sensors mount directly on your sap buckets, measuring fill levels using ultrasonic sensors and tracking temperature conditions. Each sensor operates independently with ultra-low power consumption.
                      </p>
                    </div>
                  </div>
                  <div className="relative w-full aspect-video rounded-lg overflow-hidden border border-white/10 bg-black/20">
                    <video
                      className="w-full h-full object-cover"
                      src="/sample.mp4"
                      autoPlay
                      loop
                      muted
                      playsInline
                    />
                  </div>
                </div>
              </div>
              {/* Step 02 - Video on left, text on right */}
              <div className="p-6 rounded-lg border border-white/10 bg-black/40 backdrop-blur-sm">
                <div className="grid md:grid-cols-2 gap-6 items-center">
                  <div className="relative w-full aspect-video rounded-lg overflow-hidden border border-white/10 bg-black/20 md:order-1 order-2">
                    <video
                      className="w-full h-full object-cover"
                      src="/sample.mp4"
                      autoPlay
                      loop
                      muted
                      playsInline
                    />
                  </div>
                  <div className="flex items-start gap-4 md:order-2 order-1">
                    <div className="text-2xl font-light text-white/40">02</div>
                    <div>
                      <h3 className="text-xl font-light mb-2">Wireless Data Transmission</h3>
                      <p className="text-white/60 font-light leading-relaxed">
                        Sensors transmit real-time data via Wi-Fi 6 to a Raspberry Pi controller using MQTT protocol. The controller aggregates data and securely sends it to our cloud-based database, ensuring reliable connectivity even in remote locations.
                      </p>
                    </div>
                  </div>
                </div>
              </div>
              {/* Step 03 - Video on right, text on left */}
              <div className="p-6 rounded-lg border border-white/10 bg-black/40 backdrop-blur-sm">
                <div className="grid md:grid-cols-2 gap-6 items-center">
                  <div className="flex items-start gap-4">
                    <div className="text-2xl font-light text-white/40">03</div>
                    <div>
                      <h3 className="text-xl font-light mb-2">Intelligent Analytics</h3>
                      <p className="text-white/60 font-light leading-relaxed">
                        Our backend processes sensor data alongside weather forecasts to predict optimal collection windows. The system alerts you when buckets are full, when conditions are ideal for sap flow, or if issues are detected.
                      </p>
                    </div>
                  </div>
                  <div className="relative w-full aspect-video rounded-lg overflow-hidden border border-white/10 bg-black/20">
                    <video
                      className="w-full h-full object-cover"
                      src="/sample.mp4"
                      autoPlay
                      loop
                      muted
                      playsInline
                    />
                  </div>
                </div>
              </div>
              {/* Step 04 - Video on left, text on right */}
              <div className="p-6 rounded-lg border border-white/10 bg-black/40 backdrop-blur-sm">
                <div className="grid md:grid-cols-2 gap-6 items-center">
                  <div className="relative w-full aspect-video rounded-lg overflow-hidden border border-white/10 bg-black/20 md:order-1 order-2">
                    <video
                      className="w-full h-full object-cover"
                      src="/sample.mp4"
                      autoPlay
                      loop
                      muted
                      playsInline
                    />
                  </div>
                  <div className="flex items-start gap-4 md:order-2 order-1">
                    <div className="text-2xl font-light text-white/40">04</div>
                    <div>
                      <h3 className="text-xl font-light mb-2">Interactive Dashboard</h3>
                      <p className="text-white/60 font-light leading-relaxed">
                        Access your entire operation from anywhere using our web-based dashboard. View real-time sensor data on an interactive map, track collection trends, and receive intelligent recommendations for optimal sap collection timing.
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </section>

          
          {/* Benefits Section - Placeholder for future content */}
          <section className="mb-20 max-w-4xl mx-auto relative z-30">
            <div className="text-center mb-12">
              <h2 className="text-3xl sm:text-4xl font-light mb-4">Why Choose Us?</h2>
            </div>
            <div className="space-y-6">
              <div className="p-6 rounded-lg border border-white/10 bg-black/40 backdrop-blur-sm">
                <h3 className="text-xl font-light mb-2">Designed for Small Producers</h3>
                <p className="text-white/60 font-light leading-relaxed">
                  {/* Add content about affordability and accessibility for small-scale producers */}
                </p>
              </div>
              <div className="p-6 rounded-lg border border-white/10 bg-black/40 backdrop-blur-sm">
                <h3 className="text-xl font-light mb-2">Easy Installation</h3>
                <p className="text-white/60 font-light leading-relaxed">
                  {/* Add content about simple setup and installation process */}
                </p>
              </div>
              <div className="p-6 rounded-lg border border-white/10 bg-black/40 backdrop-blur-sm">
                <h3 className="text-xl font-light mb-2">Weatherproof & Durable</h3>
                <p className="text-white/60 font-light leading-relaxed">
                  {/* Add content about durability in harsh winter conditions */}
                </p>
              </div>
            </div>
          </section>
        </div>
      </main>

      {/* Footer */}
      <footer className="relative z-10 border-t border-white/10 bg-black/60 backdrop-blur-md">
        <div className="container mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
            {/* Scroll Indicator */}
            <div className="flex items-center gap-2 text-sm text-white/60 font-light">
              <ChevronDown className="h-4 w-4" />
              <span>02/03. Scroll down</span>
            </div>

            {/* Team Credits */}
            <div className="flex items-center gap-6 flex-wrap justify-center">
              <div className="text-xs text-white/40 font-light">
                Built by: Kelsey Erb, CompE, William Neel, CompE, Aadityaa Rengaraj Sethuraman, CompE, Vansh Singh, CompE, and Keshav Verma, CompE
              </div>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}

function DataPoint({
  name,
  value,
  icon,
  description,
}: {
  name: string;
  value: string;
  icon?: "triangle" | "star" | "propeller" | "diamond";
  description?: string;
}) {
  const iconComponents = {
    triangle: (
      <div className="w-4 h-4 border-l-2 border-t-2 border-r-2 border-b-0 border-white rotate-45" />
    ),
    star: (
      <div className="relative w-4 h-4">
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="w-2 h-2 border border-white rotate-45" />
        </div>
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="w-3 h-3 border border-white rotate-0" />
        </div>
      </div>
    ),
    propeller: (
      <div className="relative w-4 h-4 flex items-center justify-center">
        <div className="w-3 h-3 border border-white rounded-full" />
        <div className="absolute w-4 h-0.5 bg-white/50" />
        <div className="absolute w-0.5 h-4 bg-white/50" />
      </div>
    ),
    diamond: (
      <div className="w-4 h-4 border border-white rotate-45" />
    ),
  };

  return (
    <div className="relative z-30 flex flex-col items-center gap-3 p-5 rounded-lg border border-white/10 hover:border-white/20 hover:bg-white/5 transition-all bg-black/40 backdrop-blur-sm">
      {icon && (
        <div className="flex items-center justify-center w-8 h-8">{iconComponents[icon]}</div>
      )}
      <div className="text-center">
        <p className="text-xs text-white/60 font-light mb-1">{name}</p>
        <p className="text-lg font-light text-white mb-2">{value}</p>
        {description && (
          <p className="text-xs text-white/40 font-light leading-relaxed max-w-[140px]">
            {description}
          </p>
        )}
      </div>
    </div>
  );
}
