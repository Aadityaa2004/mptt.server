"use client";

import { useState, useEffect, useMemo } from "react";
import { useRouter } from "next/navigation";
import { useRequireAuth } from "@/hooks/useRequireAuth";
import Navbar from "@/components/navbar/Navbar";
import { sensorService, type Pi, type PiDevice } from "@/services/api/sensorService";
import { DeviceCard } from "@/components/sensors/DeviceCard";
import { Loader2, AlertCircle, Cpu, Radio, Activity, Database } from "lucide-react";

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

  if (isLoading || isLoadingSensors) {
    return (
      <div className="min-h-screen bg-black text-white flex items-center justify-center">
        <div className="flex flex-col items-center gap-3">
          <Loader2 className="h-6 w-6 text-white/60 animate-spin" />
          <p className="text-white/60 font-light">Loading sensors...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-black text-white">
      <Navbar />
      <main className="pt-24 px-4 sm:px-6 lg:px-8 pb-16">
        <div className="max-w-7xl mx-auto">
          {/* Header Section */}
          <div className="mb-8">
            <h1 className="text-4xl font-light tracking-tight mb-2">
              My Sensors
            </h1>
            {/* <p className="text-white/60 font-light text-sm">
              View and manage your sensor devices
            </p> */}
          </div>

          {/* Summary Statistics */}
          {pis.length > 0 && (
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-8">
              <div className="border border-white/10 rounded-lg p-4 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm">
                <div className="flex items-center gap-3">
                  <div>
                    <p className="text-white/60 font-light text-xs">Total Devices</p>
                    <p className="text-xl font-light">{stats.totalDevices}</p>
                  </div>
                </div>
              </div>
              <div className="border border-white/10 rounded-lg p-4 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm">
                <div className="flex items-center gap-3">
                  <div>
                    <p className="text-white/60 font-light text-xs">Raspberry Pi Units</p>
                    <p className="text-xl font-light">{stats.totalPis}</p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {error && (
            <div className="mb-6 p-4 border border-red-500/20 bg-red-500/10 rounded-lg flex items-center gap-2">
              <AlertCircle className="h-5 w-5 text-red-400" />
              <p className="text-sm text-red-400 font-light">{error}</p>
            </div>
          )}

          {pis.length === 0 && !isLoadingSensors ? (
            <div className="border border-white/10 rounded-lg p-12 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm text-center">
              <Radio className="h-12 w-12 text-white/40 mx-auto mb-4" />
              <p className="text-white/60 font-light mb-2">No sensors found</p>
              <p className="text-white/40 font-light text-sm">
                Your sensors will appear here once they are registered
              </p>
            </div>
          ) : (
            <div className="space-y-6">
              {pis.map((pi) => (
                <div
                  key={pi.pi_id}
                  className="border border-white/10 rounded-lg bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm overflow-hidden transition-all duration-200 hover:border-white/15"
                >
                  {/* PI Header */}
                  <div className="px-6 py-4 border-b border-white/10 bg-white/5">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <div>
                          <h2 className="text-xl font-light">Raspberry Pi Unit: <span className="text-white/60 font-bold text-sm bg-orange-500/80 px-2 py-1 rounded-md">{pi.pi_id}</span></h2>
                          <p className="text-white/60 font-light text-sm">
                            {pi.devices.length} device{pi.devices.length !== 1 ? "s" : ""} connected
                          </p>
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Devices Grid */}
                  {pi.devices.length > 0 ? (
                    <div className="p-6">
                      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 items-stretch">
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
                    </div>
                  ) : (
                    <div className="p-6 text-center">
                      <p className="text-white/40 font-light text-sm">
                        No devices connected to this PI
                      </p>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </main>
    </div>
  );
}

