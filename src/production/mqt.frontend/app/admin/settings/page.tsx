"use client";

export default function AdminSettingsPage() {
  return (
    <main className="pt-16 sm:pt-24 px-4 sm:px-6 lg:px-8 pb-12 sm:pb-16">
      <div className="max-w-4xl mx-auto">
        <div className="mb-8">
          <h1 className="text-2xl sm:text-4xl font-light tracking-tight mb-2">Settings</h1>
          <p className="text-muted-foreground font-light text-sm">
            Manage admin settings and preferences
          </p>
        </div>

        <div className="border border-border rounded-lg p-8 bg-card">
          <h2 className="text-2xl font-light mb-2">Administration</h2>
          <p className="text-muted-foreground font-light text-sm">
            Admin-specific settings can be configured here.
          </p>
        </div>
      </div>
    </main>
  );
}
