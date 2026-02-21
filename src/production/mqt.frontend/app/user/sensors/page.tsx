"use client";

import { useState, useEffect, useMemo } from "react";
import { useRouter } from "next/navigation";
import { useRequireAuth } from "@/hooks/useRequireAuth";
import Navbar from "@/components/navbar/Navbar";
import { sensorService, type Pi, type PiDevice } from "@/services/api/sensorService";
import { DeviceCard } from "@/components/sensors/DeviceCard";
import { Loader2, AlertCircle, Cpu, Radio, RefreshCw } from "lucide-react";

interface PiWithDevices extends Pi {
  devices: PiDevice[];
}

export default function MySensorsPage() {
  const { user, isLoading } = useRequireAuth("user");
  const router = useRouter();
  const [pis, setPis] = useState<PiWithDevices[]>([]);
  const [isLoadingSensors, setIsLoadingSensors] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!isLoading && user) {
      loadSensors();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isLoading, user]);

  const loadSensors = async () => {
    if (!user?.user_id) return;

    try {
      setIsLoadingSensors(true);
      setError(null);

      // Fetch all PIs for the user
      const pisResponse = await sensorService.getPis({
        user_id: user.user_id,
        page: 1,
        page_size: 100, // Get all PIs
      });

      // Fetch devices for each PI
      const pisWithDevices = await Promise.all(
        pisResponse.items.map(async (pi) => {
          try {
            const devicesResponse = await sensorService.getDevices({
              pi_id: pi.pi_id,
              page: 1,
              page_size: 100, // Get all devices
            });
            return {
              ...pi,
              devices: devicesResponse.items,
            };
          } catch (err) {
            console.error(`Error loading devices for PI ${pi.pi_id}:`, err);
            return {
              ...pi,
              devices: [],
            };
          }
        })
      );

      setPis(pisWithDevices);
    } catch (err) {
      console.error("Error loading sensors:", err);
      const errorMessage = err instanceof Error ? err.message : "Failed to load sensors";
      setError(errorMessage);
    } finally {
      setIsLoadingSensors(false);
    }
  };

  // Calculate summary statistics
  const stats = useMemo(() => {
    const totalDevices = pis.reduce((sum, pi) => sum + pi.devices.length, 0);
    const totalPis = pis.length;
    return {
      totalDevices,
      totalPis,
    };
  }, [pis]);

  if (isLoading || (isLoadingSensors && pis.length === 0)) {
    return (
      <div className="min-h-screen bg-background text-foreground flex items-center justify-center">
        <div className="flex flex-col items-center gap-3">
          <Loader2 className="h-6 w-6 text-white/60 animate-spin" />
          <p className="text-white/60 font-light">Loading sensors...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background text-foreground">
      <Navbar />
      <main className="pt-16 sm:pt-20">
        <div className="max-w-[1600px] mx-auto">
          {/* Header */}
          <div className="px-4 sm:px-6 lg:px-8 py-4 sm:py-6 border-b border-white/5 flex flex-col sm:flex-row sm:items-start sm:justify-between gap-3">
            <div>
              <h1 className="text-2xl sm:text-3xl font-light tracking-tight text-foreground">My Sensors</h1>
              {pis.length > 0 && (
                <p className="text-white/50 font-light text-sm mt-1">
                  {stats.totalDevices} device{stats.totalDevices !== 1 ? "s" : ""} across {stats.totalPis} PI{stats.totalPis !== 1 ? "s" : ""}
                </p>
              )}
            </div>
            <button
              onClick={() => loadSensors()}
              disabled={isLoadingSensors}
              className="p-2.5 rounded-xl bg-white/5 hover:bg-white/10 text-white/70 hover:text-white disabled:opacity-50 transition-all"
              title="Refresh sensors"
            >
              <RefreshCw className={`h-5 w-5 ${isLoadingSensors ? "animate-spin" : ""}`} />
            </button>
          </div>

          {error && (
            <div className="mx-4 sm:mx-6 lg:mx-8 mt-4 p-4 rounded-xl bg-red-500/5 border border-red-500/20 flex items-center gap-2">
              <AlertCircle className="h-5 w-5 text-red-400 flex-shrink-0" />
              <p className="text-sm text-red-400 font-light">{error}</p>
            </div>
          )}

          {pis.length === 0 && !isLoadingSensors ? (
            <section className="px-4 sm:px-6 lg:px-8 py-16 text-center">
              <div className="max-w-md mx-auto">
                <div className="w-16 h-16 rounded-2xl bg-white/5 flex items-center justify-center mx-auto mb-4">
                  <Radio className="h-8 w-8 text-white/40" />
                </div>
                <p className="text-white/70 font-light mb-2">No sensors yet</p>
                <p className="text-white/45 font-light text-sm">
                  Your sensors will appear here once they are registered and connected.
                </p>
              </div>
            </section>
          ) : (
            <div className="px-4 sm:px-6 lg:px-8 py-6 space-y-8">
              {pis.map((pi) => (
                <section key={pi.pi_id}>
                  <div className="flex items-center gap-3 mb-4">
                    <div className="w-9 h-9 rounded-xl bg-orange-500/20 flex items-center justify-center">
                      <Cpu className="h-4 w-4 text-orange-400/90" />
                    </div>
                    <div>
                      <h2 className="text-lg font-light text-foreground">{pi.pi_id}</h2>
                      <p className="text-xs text-white/50 font-light">
                        {pi.devices.length} device{pi.devices.length !== 1 ? "s" : ""}
                      </p>
                    </div>
                  </div>
                  {pi.devices.length > 0 ? (
                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                      {pi.devices.map((device) => (
                        <DeviceCard
                          key={device.device_id}
                          device={{
                            pi_id: pi.pi_id,
                            device_id: device.device_id,
                          }}
                        />
                      ))}
                    </div>
                  ) : (
                    <div className="py-8 rounded-xl bg-white/[0.02] border border-white/5 text-center">
                      <p className="text-white/45 font-light text-sm">No devices connected to this PI</p>
                    </div>
                  )}
                </section>
              ))}
            </div>
          )}
        </div>
      </main>
    </div>
  );
}

