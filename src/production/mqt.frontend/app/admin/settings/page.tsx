"use client";

import { RefreshCw } from "lucide-react";
import { ThemeToggle } from "@/components/ThemeToggle";

export default function AdminSettingsPage() {
  return (
    <main className="pt-24 px-4 sm:px-6 lg:px-8 pb-16">
      <div className="max-w-4xl mx-auto">
        <div className="mb-8 flex items-start justify-between">
          <div>
            <h1 className="text-4xl font-light tracking-tight mb-2">Settings</h1>
            <p className="text-muted-foreground font-light text-sm">
              Manage admin settings and preferences
            </p>
          </div>
          <button
            onClick={() => window.location.reload()}
            className="p-2 rounded-lg border border-white/20 hover:bg-white/5 text-white/70 hover:text-white transition-colors"
            title="Refresh"
          >
            <RefreshCw className="h-4 w-4" />
          </button>
        </div>

        <div className="border border-border rounded-lg p-8 bg-card">
          <h2 className="text-2xl font-light mb-2">Appearance</h2>
          <p className="text-muted-foreground font-light text-sm mb-4">
            Choose light or dark theme for the application.
          </p>
          <ThemeToggle />
        </div>
      </div>
    </main>
  );
}
