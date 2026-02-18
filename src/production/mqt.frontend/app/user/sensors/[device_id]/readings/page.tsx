"use client";

import { useState, useEffect } from "react";
import { useRouter, useParams } from "next/navigation";
import { useRequireAuth } from "@/hooks/useRequireAuth";
import Navbar from "@/components/navbar/Navbar";
import { sensorService } from "@/services/api/sensorService";
import { ReadingsTable } from "@/components/sensors/ReadingsTable";
import { Loader2, AlertCircle, ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";

export default function SensorReadingsPage() {
  const { user, isLoading } = useRequireAuth("user");
  const router = useRouter();
  const params = useParams();
  const deviceId = params?.device_id as string;

  const [piId, setPiId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!isLoading && user && deviceId) {
      findPiForDevice();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isLoading, user, deviceId]);

  const findPiForDevice = async () => {
    if (!user?.user_id) return;

    try {
      // Fetch all PIs for the user
      const pisResponse = await sensorService.getPis({
        user_id: user.user_id,
        page: 1,
        page_size: 100,
      });

      // Find the PI that contains this device
      for (const pi of pisResponse.items) {
        try {
          const devicesResponse = await sensorService.getDevices({
            pi_id: pi.pi_id,
            page: 1,
            page_size: 100,
          });

          const device = devicesResponse.items.find((d) => d.device_id === deviceId);
          if (device) {
            setPiId(pi.pi_id);
            return;
          }
        } catch (err) {
          console.error(`Error checking devices for PI ${pi.pi_id}:`, err);
        }
      }

      setError("Device not found or you don't have access to it");
    } catch (err) {
      console.error("Error finding PI for device:", err);
      const errorMessage = err instanceof Error ? err.message : "Failed to find device";
      setError(errorMessage);
    }
  };

  if (isLoading || !piId) {
    return (
      <div className="min-h-screen bg-black text-white flex items-center justify-center">
        <div className="flex flex-col items-center gap-3">
          <Loader2 className="h-6 w-6 text-white/60 animate-spin" />
          <p className="text-white/60 font-light">Loading sensor data...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-black text-white">
      <Navbar />
      <main className="pt-24 px-4 sm:px-6 lg:px-8 pb-16">
        <div className="max-w-7xl mx-auto">
          {/* Header */}
          <div className="mb-8">
            <Button
              variant="ghost"
              onClick={() => router.push(`/user/sensors/${deviceId}`)}
              className="mb-4 text-white/60 hover:text-white"
            >
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back to Analytics
            </Button>
            <h1 className="text-4xl font-light tracking-tight mb-2">
              Sensor Readings
            </h1>
            <p className="text-white/60 font-light text-sm font-mono">
              Device ID: {deviceId}
            </p>
            <p className="text-white/60 font-light text-sm">
              PI ID: {piId}
            </p>
          </div>

          {error && (
            <div className="mb-6 p-4 border border-red-500/20 bg-red-500/10 rounded-lg flex items-center gap-2">
              <AlertCircle className="h-5 w-5 text-red-400" />
              <p className="text-sm text-red-400 font-light">{error}</p>
            </div>
          )}

          {/* Readings Table */}
          <div className="border border-white/10 rounded-lg bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm overflow-hidden">
            <div className="px-6 py-4 border-b border-white/10">
              <h2 className="text-xl font-light">Readings History</h2>
            </div>
            {piId && <ReadingsTable deviceId={deviceId} piId={piId} />}
          </div>
        </div>
      </main>
    </div>
  );
}

