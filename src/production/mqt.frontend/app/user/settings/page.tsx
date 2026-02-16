"use client";

import { useState, useEffect } from "react";
import { useRequireAuth } from "@/hooks/useRequireAuth";
import Navbar from "@/components/navbar/Navbar";
import { LocationInput } from "@/components/weather/LocationInput";
import { weatherService } from "@/services/api/weatherService";
import { sensorService, type Pi } from "@/services/api/sensorService";
import { usePiPreferences, colorToGradient } from "@/hooks/usePiPreferences";
import { MarkerShapeComponent } from "@/components/map/MarkerShape";
import { Loader2, AlertCircle, CheckCircle } from "lucide-react";


export default function SettingsPage() {
  const { user, isLoading } = useRequireAuth("user");
  const { preferences, getPreference, updatePreference, initializePreferences, loadColorsFromBackend, isLoadingFromBackend } = usePiPreferences();
  const [hasLocation, setHasLocation] = useState<boolean | null>(null);
  const [isCheckingLocation, setIsCheckingLocation] = useState(true);
  const [isUpdatingLocation, setIsUpdatingLocation] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [locationName, setLocationName] = useState<string>("");
  const [pis, setPis] = useState<Pi[]>([]);
  const [isLoadingPis, setIsLoadingPis] = useState(false);
  const [openColorPickers, setOpenColorPickers] = useState<Record<string, boolean>>({});
  const [customColors, setCustomColors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (!isLoading && user) {
      checkLocation();
      loadPis();
    }
  }, [isLoading, user]);

  const loadPis = async () => {
    if (!user?.user_id) return;

    try {
      setIsLoadingPis(true);
      const pisResponse = await sensorService.getPis({
        user_id: user.user_id,
        page: 1,
        page_size: 100,
      });
      const loadedPis = Array.isArray(pisResponse?.items) ? pisResponse.items : [];
      setPis(loadedPis);
      
      // Initialize preferences for all PIs
      if (loadedPis.length > 0) {
        initializePreferences(loadedPis.map(pi => pi.pi_id));
        // Load colors from backend device locations
        await loadColorsFromBackend();
      }
    } catch (err) {
      console.error("Error loading PIs:", err);
    } finally {
      setIsLoadingPis(false);
    }
  };

  const checkLocation = async () => {
    try {
      setIsCheckingLocation(true);
      const profile = await weatherService.getProfile();
      if (profile && profile.latitude !== null && profile.latitude !== undefined && 
          profile.longitude !== null && profile.longitude !== undefined) {
        setHasLocation(true);
      } else {
        setHasLocation(false);
      }
    } catch (err) {
      console.error("Error checking location:", err);
      setError("Failed to check location status");
      setHasLocation(false);
    } finally {
      setIsCheckingLocation(false);
    }
  };

  const handleLocationSubmit = async (latitude: number, longitude: number, name: string) => {
    try {
      setIsUpdatingLocation(true);
      setError(null);
      setSuccess(null);
      await weatherService.updateLocation({ latitude, longitude });
      setLocationName(name);
      setHasLocation(true);
      setSuccess("Location updated successfully!");
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      console.error("Error updating location:", err);
      const errorMessage = err instanceof Error ? err.message : "Failed to update location";
      setError(errorMessage);
    } finally {
      setIsUpdatingLocation(false);
    }
  };

  if (isLoading || isCheckingLocation) {
    return (
      <div className="min-h-screen bg-black text-white flex items-center justify-center">
        <div className="flex flex-col items-center gap-3">
          <Loader2 className="h-6 w-6 text-white/60 animate-spin" />
          <p className="text-white/60 font-light">Loading...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-black text-white">
      <Navbar />
      <main className="pt-24 px-4 sm:px-6 lg:px-8 pb-16">
        <div className="max-w-4xl mx-auto">
          <div className="mb-8">
            <h1 className="text-4xl font-light tracking-tight mb-2">
              Settings
            </h1>
            <p className="text-white/60 font-light text-sm">
              Manage your account settings and preferences
            </p>
          </div>

          {error && (
            <div className="mb-6 p-4 border border-red-500/20 bg-red-500/10 rounded-lg flex items-center gap-2">
              <AlertCircle className="h-5 w-5 text-red-400" />
              <p className="text-sm text-red-400 font-light">{error}</p>
            </div>
          )}

          {success && (
            <div className="mb-6 p-4 border border-green-500/20 bg-green-500/10 rounded-lg flex items-center gap-2">
              <CheckCircle className="h-5 w-5 text-green-400" />
              <p className="text-sm text-green-400 font-light">{success}</p>
            </div>
          )}

          {/* Location Settings */}
          <div className="relative z-10 border border-white/10 rounded-lg p-8 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm mb-6">
            <h2 className="text-2xl font-light mb-4">Location Settings</h2>
            <p className="text-white/60 font-light text-sm mb-6">
              {hasLocation
                ? `Current location: ${locationName || "Location set"}. Update your location below.`
                : "Set your location to enable weather features and accurate sensor mapping."}
            </p>
            <LocationInput
              onLocationSubmit={handleLocationSubmit}
              isLoading={isUpdatingLocation}
            />
          </div>

          {/* PI Marker Settings */}
          <div className="border border-white/10 rounded-lg p-8 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm">
            <h2 className="text-2xl font-light mb-4">Map Marker Settings</h2>
            <p className="text-white/60 font-light text-sm mb-6">
              Customize the colors of markers for each Raspberry Pi unit on the map. This helps you visually group and identify your devices.
            </p>

            {isLoadingPis ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-6 w-6 text-white/60 animate-spin" />
              </div>
            ) : pis.length === 0 ? (
              <p className="text-white/60 font-light text-sm text-center py-8">
                No Raspberry Pi units found. Add devices to your account to configure marker settings.
              </p>
            ) : (
              <div className="space-y-3">
                {pis.map((pi, index) => {
                  const preference = getPreference(pi.pi_id, index);
                  return (
                    <div
                      key={pi.pi_id}
                      className="flex items-center gap-4 p-4 border border-white/10 rounded-lg bg-white/5 hover:bg-white/10 transition-colors"
                    >
                      {/* PI Name */}
                      <div className="flex-1 min-w-0">
                        <h3 className="text-base font-light">
                          <span className="text-white/60">PI:</span>{" "}
                          <span className="text-white/90 font-mono">{pi.pi_id}</span>
                        </h3>
                      </div>

                      {/* Current Color Preview */}
                      <div className="flex items-center gap-2">
                        <MarkerShapeComponent
                          gradient={colorToGradient(preference.color)}
                          size="sm"
                        />
                      </div>

                      {/* Custom Color Picker */}
                      <div className="flex-shrink-0 relative">
                        <button
                          type="button"
                          onClick={() => {
                            setOpenColorPickers(prev => ({ ...prev, [pi.pi_id]: !prev[pi.pi_id] }));
                            // Initialize custom color from current preference if not set
                            if (!customColors[pi.pi_id]) {
                              setCustomColors(prev => ({
                                ...prev,
                                [pi.pi_id]: preference.color || "#f97316",
                              }));
                            }
                          }}
                          disabled={isLoadingFromBackend}
                          className="flex h-10 items-center justify-center rounded-md border border-white/10 bg-white/5 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-orange-500/50 focus:border-orange-500/50 disabled:cursor-not-allowed disabled:opacity-50 hover:bg-white/10 transition-colors"
                          title="Customize gradient colors"
                        >
                          <svg className="w-5 h-5 text-white/60" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01" />
                          </svg>
                        </button>
                        
                        {/* Color Picker Panel */}
                        {openColorPickers[pi.pi_id] && (
                          <>
                            <div
                              className="fixed inset-0 z-10"
                              onClick={() => setOpenColorPickers(prev => ({ ...prev, [pi.pi_id]: false }))}
                            />
                            <div className="absolute z-20 mt-2 right-0 bg-black/95 border border-white/10 rounded-lg shadow-lg backdrop-blur-md p-4 min-w-[240px]">
                              <h4 className="text-sm font-light mb-3 text-white">Marker Color</h4>
                              
                              {/* Color Picker */}
                              <div className="mb-4">
                                <label className="text-xs text-white/60 font-light mb-2 block">Color</label>
                                <div className="flex items-center gap-3">
                                  <input
                                    type="color"
                                    value={customColors[pi.pi_id] || "#f97316"}
                                    onChange={(e) => {
                                      setCustomColors(prev => ({
                                        ...prev,
                                        [pi.pi_id]: e.target.value,
                                      }));
                                    }}
                                    className="w-12 h-10 rounded border border-white/10 cursor-pointer"
                                  />
                                  <div className="flex-1">
                                    <input
                                      type="text"
                                      value={customColors[pi.pi_id] || "#f97316"}
                                      onChange={(e) => {
                                        if (/^#[0-9A-Fa-f]{6}$/.test(e.target.value)) {
                                          setCustomColors(prev => ({
                                            ...prev,
                                            [pi.pi_id]: e.target.value,
                                          }));
                                        }
                                      }}
                                      className="w-full h-10 rounded-md border border-white/10 bg-white/5 px-3 py-2 text-sm text-white font-mono focus:outline-none focus:ring-2 focus:ring-orange-500/50 focus:border-orange-500/50"
                                      placeholder="#f97316"
                                    />
                                  </div>
                                </div>
                              </div>

                              {/* Preview */}
                              <div className="mb-4 flex items-center gap-3">
                                <span className="text-xs text-white/60 font-light">Preview:</span>
                                <MarkerShapeComponent
                                  gradient={`from-[${customColors[pi.pi_id] || "#f97316"}] to-[${customColors[pi.pi_id] || "#f97316"}]`}
                                  size="sm"
                                />
                              </div>

                              {/* Actions */}
                              <div className="flex gap-2">
                                <button
                                  type="button"
                                  onClick={async () => {
                                    const colorHex = customColors[pi.pi_id] || "#f97316";
                                    
                                    setOpenColorPickers(prev => ({ ...prev, [pi.pi_id]: false }));
                                    try {
                                      await updatePreference(pi.pi_id, { color: colorHex });
                                      setSuccess(`Color updated for ${pi.pi_id}`);
                                      setTimeout(() => setSuccess(null), 2000);
                                    } catch (err) {
                                      console.error("Error updating color:", err);
                                      setError("Failed to sync color to backend. Changes saved locally.");
                                      setTimeout(() => setError(null), 3000);
                                    }
                                  }}
                                  className="flex-1 h-9 rounded-md bg-white/10 hover:bg-white/20 border border-white/20 text-white text-sm font-light transition-colors"
                                >
                                  Apply
                                </button>
                                <button
                                  type="button"
                                  onClick={() => setOpenColorPickers(prev => ({ ...prev, [pi.pi_id]: false }))}
                                  className="h-9 px-4 rounded-md border border-white/20 text-white/60 hover:text-white hover:bg-white/10 text-sm font-light transition-colors"
                                >
                                  Cancel
                                </button>
                              </div>
                            </div>
                          </>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}

