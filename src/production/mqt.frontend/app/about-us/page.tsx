"use client";

import Navbar from "@/components/navbar/Navbar";
import { DottedGlowBackground } from "@/components/ui/dotted-glow-background";
import { GraduationCap, Users, Target } from "lucide-react";

export default function AboutUsPage() {
  const teamMembers = [
    {
      name: "Kelsey Erb",
      role: "Computer Engineering",
      initials: "KE",
      image: "/persom.jpg",
      bio: "Focuses on hardware integration and sensor optimization.",
    },
    {
      name: "William Neel",
      role: "Computer Engineering",
      initials: "WN",
      image: "/persom.jpg",
      bio: "Leads backend architecture and API design.",
    },
    {
      name: "Aadityaa R. Sethuraman",
      role: "Computer Engineering",
      initials: "AS",
      image: "/persom.jpg",
      bio: "Creates intuitive interfaces for maple producers.",
    },
    {
      name: "Vansh Singh",
      role: "Computer Engineering",
      initials: "VS",
      image: "/persom.jpg",
      bio: "Develops predictive models for sap collection.",
    },
    {
      name: "Keshav Verma",
      role: "Computer Engineering",
      initials: "KV",
      image: "/persom.jpg",
      bio: "Ensures robust device communication and reliability.",
    },
  ];

  return (
    <div className="relative min-h-screen bg-black text-white overflow-hidden">
      <Navbar />

      {/* Main Content */}
      <main className="relative z-20 px-4 sm:px-6 lg:px-8 pt-24 pb-20">
        <div className="container mx-auto max-w-6xl">
          {/* Hero Section */}
          <div className="relative mb-20 pt-16 pb-16 min-h-[40vh] rounded-2xl overflow-hidden">
            {/* Background Effects */}
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
            <div className="relative z-10 text-center">
              <h1 className="text-4xl sm:text-5xl lg:text-6xl font-light tracking-tight mb-4">
                About MapleSense
              </h1>
              <p className="text-lg sm:text-xl text-white/70 font-light max-w-2xl mx-auto">
                Empowering small-scale maple producers with innovative technology
              </p>
            </div>
          </div>

          {/* Vision 2027 Section */}
          <section className="mb-20 max-w-4xl mx-auto relative z-30">
            <div className="flex items-center gap-4 mb-8">
              {/* <div className="p-3 rounded-lg border border-white/10 bg-white/5">
                <Target className="h-6 w-6 text-white/70" />
              </div> */}
              <div>
                <h2 className="text-3xl sm:text-4xl font-light mb-2">Vision 2027</h2>
                {/* <p className="text-sm text-white/50 font-light">Our vision for the future</p> */}
              </div>
            </div>

            <div className="p-8 rounded-lg border border-white/10 bg-black/40 backdrop-blur-sm">
              <p className="text-lg text-white/80 font-light leading-relaxed mb-6">
                By 2027, MapleSense aims to become the leading monitoring solution for small-scale and hobbyist maple syrup producers across North America. We envision a future where traditional bucket and tap setups are seamlessly integrated with modern IoT technology, making sap collection more efficient, accessible, and sustainable.
              </p>
              <div className="grid md:grid-cols-2 gap-6 mt-8">
                <div className="p-5 rounded-lg border border-white/10 bg-white/5">
                  <h3 className="text-lg font-light mb-2 text-white/90">Accessibility</h3>
                  <p className="text-sm text-white/60 font-light leading-relaxed">
                    Making advanced monitoring technology affordable and accessible to producers of all scales, ensuring that small operations can compete and thrive.
                  </p>
                </div>
                <div className="p-5 rounded-lg border border-white/10 bg-white/5">
                  <h3 className="text-lg font-light mb-2 text-white/90">Innovation</h3>
                  <p className="text-sm text-white/60 font-light leading-relaxed">
                    Continuously evolving our platform with cutting-edge sensors, AI-powered insights, and predictive analytics to optimize sap collection yields.
                  </p>
                </div>
                <div className="p-5 rounded-lg border border-white/10 bg-white/5">
                  <h3 className="text-lg font-light mb-2 text-white/90">Sustainability</h3>
                  <p className="text-sm text-white/60 font-light leading-relaxed">
                    Supporting sustainable maple syrup production through efficient resource management and reduced waste, preserving traditional methods while embracing innovation.
                  </p>
                </div>
                <div className="p-5 rounded-lg border border-white/10 bg-white/5">
                  <h3 className="text-lg font-light mb-2 text-white/90">Community</h3>
                  <p className="text-sm text-white/60 font-light leading-relaxed">
                    Building a vibrant community of maple producers who share knowledge, best practices, and experiences to collectively advance the industry.
                  </p>
                </div>
              </div>
            </div>
          </section>

          {/* Team Section */}
          <section className="mb-20 max-w-4xl mx-auto relative z-30">
            <div className="flex items-center gap-4 mb-8">
              {/* <div className="p-3 rounded-lg border border-white/10 bg-white/5">
                <Users className="h-6 w-6 text-white/70" />
              </div> */}
              <div>
                <h2 className="text-3xl sm:text-4xl font-light mb-2">Our Team</h2>
                {/* <p className="text-sm text-white/50 font-light">Meet the engineers behind MapleSense</p> */}
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
              {teamMembers.map((member, index) => (
                <div
                  key={index}
                  className="group relative overflow-hidden rounded-md border border-white/10 bg-black/40 backdrop-blur-sm hover:border-white/20 hover:bg-white/5 transition-all duration-300"
                >
                  <div className="flex flex-col p-4">
                    {/* Profile Image */}
                    <div className="relative mb-4 flex justify-center">
                      <div className="relative w-full aspect-square max-w-64 rounded-md overflow-hidden border border-white/20 bg-gradient-to-br from-orange-500/20 to-white/10 group-hover:border-white/30 transition-all">
                        <img
                          src={member.image}
                          alt={member.name}
                          className="w-full h-full object-cover opacity-90 group-hover:opacity-100 transition-opacity"
                          onError={(e) => {
                            // Fallback to initials if image fails to load
                            const target = e.target as HTMLImageElement;
                            target.style.display = 'none';
                            const parent = target.parentElement;
                            if (parent) {
                              parent.innerHTML = `<span class="text-5xl font-light text-white/80 flex items-center justify-center w-full h-full">${member.initials}</span>`;
                            }
                          }}
                        />
                      </div>
                    </div>
                    
                    {/* Content */}
                    <div className="flex flex-col items-center text-center">
                      {/* Name */}
                      <h3 className="text-lg font-light text-white">
                        {member.name}
                      </h3>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </section>

          {/* Advisor Section */}
          <section className="mb-20 max-w-4xl mx-auto relative z-30">
            <div className="flex items-center gap-4 mb-8">
              {/* <div className="p-3 rounded-lg border border-white/10 bg-white/5">
                <GraduationCap className="h-6 w-6 text-white/70" />
              </div> */}
              <div>
                <h2 className="text-3xl sm:text-4xl font-light mb-2">Our Advisor</h2>
                {/* <p className="text-sm text-white/50 font-light">Guiding our journey</p> */}
              </div>
            </div>

            <div className="p-8 rounded-lg border border-white/10 bg-black/40 backdrop-blur-sm">
              <div className="flex flex-col md:flex-row items-center md:items-start gap-6">
                {/* Avatar */}
                <div className="w-24 h-24 rounded-md bg-gradient-to-br from-orange-500/20 to-white/10 border border-white/20 flex items-center justify-center flex-shrink-0">
                  <span className="text-3xl font-light text-white/80">MZ</span>
                </div>
                {/* Info */}
                <div className="flex-1 text-center md:text-left">
                  <h3 className="text-2xl font-light text-white mb-2">Professor Michael Zink</h3>
                  <p className="text-base text-white/70 font-light mb-4">
                    Faculty Advisor
                  </p>
                  <p className="text-sm text-white/60 font-light leading-relaxed">
                    We extend our deepest gratitude to Professor Michael Zink for his invaluable guidance, support, and expertise throughout the development of MapleSense. His mentorship has been instrumental in shaping our project from concept to reality, helping us navigate technical challenges and providing insights that have significantly enhanced our system design and implementation.
                  </p>
                </div>
              </div>
            </div>
          </section>

          {/* Mission Statement */}
          <section className="max-w-4xl mx-auto relative z-30">
            <div className="p-8 rounded-lg border border-white/10 bg-black/40 backdrop-blur-sm">
              <h2 className="text-2xl font-light mb-4 text-center">Our Mission</h2>
              <p className="text-lg text-white/70 font-light leading-relaxed text-center max-w-3xl mx-auto">
                MapleSense is dedicated to revolutionizing maple sap collection for small-scale and hobbyist producers. We combine cutting-edge IoT technology with traditional bucket and tap methods to create an affordable, efficient, and user-friendly monitoring solution that empowers producers to optimize their operations and maximize their yields.
              </p>
            </div>
          </section>
        </div>
      </main>

      {/* Footer */}
      <footer className="relative z-10 border-t border-white/10 bg-black/60 backdrop-blur-md mt-20">
        <div className="container mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
            <div className="text-xs text-white/40 font-light">
              Built by: Kelsey Erb, CompE, William Neel, CompE, Aadityaa Rengaraj Sethuraman, CompE, Vansh Singh, CompE, and Keshav Verma, CompE
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}

