"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { sensorService } from "@/services/api/sensorService";
import type { Reading } from "@/types/admin";
import { Radio, ChevronRight, Thermometer, Droplets, Battery, Loader2 } from "lucide-react";

interface DeviceCardProps {
  device: {
    pi_id: string;
    device_id: string;
  };
}

export function DeviceCard({ device }: DeviceCardProps) {
  const router = useRouter();
  const [latestReading, setLatestReading] = useState<Reading | null>(null);
  const [isLoadingReading, setIsLoadingReading] = useState(true);

  useEffect(() => {
    loadLatestReading();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [device.pi_id, device.device_id]);

  const loadLatestReading = async () => {
    try {
      setIsLoadingReading(true);
      const reading = await sensorService.getLatestDeviceReading(device.pi_id, device.device_id);
      setLatestReading(reading);
    } catch (err) {
      console.error("Error loading latest reading:", err);
      setLatestReading(null);
    } finally {
      setIsLoadingReading(false);
    }
  };

  const handleClick = () => {
    // Pass pi_id as query parameter to avoid having to look it up
    router.push(`/user/sensors/${device.device_id}?pi_id=${encodeURIComponent(device.pi_id)}`);
  };

  return (
    <button
      onClick={handleClick}
      className="group relative flex flex-col p-6 border border-white/10 rounded-lg bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm hover:border-white/20 hover:bg-white/5 transition-all duration-200 text-left w-full h-full"
    >
      {/* Header Section */}
      <div className="mb-4">
        <h3 className="text-lg font-light mb-1">Device</h3>
        <p className="text-white/60 font-light text-sm font-mono tracking-wider">
          {device.device_id}
        </p>
      </div>

      {/* Content Section - Flex grow to push chevron down */}
      <div className="flex-1 flex flex-col">
        {/* Latest Reading */}
        {isLoadingReading ? (
          <div className="mt-4 pt-4 border-t border-white/10 flex items-center justify-center py-2">
            <Loader2 className="h-4 w-4 text-white/40 animate-spin" />
          </div>
        ) : latestReading ? (
          <div className="mt-4 pt-4 border-t border-white/10 space-y-3">
            {/* Sensor Readings Grid */}
            <div className="grid grid-cols-3 gap-2">
              {latestReading.payload.sensors.temperature && (
                <div className="flex flex-col items-center gap-1 p-2 rounded bg-orange-500/60">
                  <Thermometer className="h-3.5 w-3.5 text-white/60" />
                  <span className="text-xs font-light text-white/90">
                    {latestReading.payload.sensors.temperature.value.toFixed(1)}°
                    {latestReading.payload.sensors.temperature.unit === "fahrenheit" || latestReading.payload.sensors.temperature.unit === "F" 
                      ? "F" 
                      : latestReading.payload.sensors.temperature.unit === "celsius" || latestReading.payload.sensors.temperature.unit === "C" 
                      ? "C" 
                      : latestReading.payload.sensors.temperature.unit?.toUpperCase() || ""}
                  </span>
                </div>
              )}
              {latestReading.payload.sensors.level && (
                <div className="flex flex-col items-center gap-1 p-2 rounded bg-orange-500/60">
                  <Droplets className="h-3.5 w-3.5 text-white/60" />
                  <span className="text-xs font-light text-white/90">
                    {latestReading.payload.sensors.level.value.toFixed(1)}
                  </span>
                  <span className="text-[10px] font-light text-white/50">
                    {latestReading.payload.sensors.level.unit}
                  </span>
                </div>
              )}
              <div className="flex flex-col items-center gap-1 p-2 rounded bg-orange-500/60">
                <Battery className="h-3.5 w-3.5 text-white/60" />
                <span className="text-xs font-light text-white/90">
                  {latestReading.payload.battery_percentage.toFixed(0)}%
                </span>
              </div>
            </div>
            
            {/* Last Updated */}
            <div className="pt-2 border-t border-white/5">
              <p className="text-xs text-white/40 font-light">
                Last updated: {new Date(latestReading.ts).toLocaleString()}
              </p>
            </div>
          </div>
        ) : (
          <div className="mt-4 pt-4 border-t border-white/10">
            <p className="text-xs text-white/40 font-light text-center">
              No readings available yet
            </p>
          </div>
        )}
      </div>

      {/* Bottom Chevron - Always at bottom */}
      <div className="mt-4 pt-4 border-t border-white/10 flex items-center justify-end">
        <div className="bg-gray-200/60 px-10 py-1 rounded-md group-hover:bg-orange-500/90 transition-colors">
          <ChevronRight className="h-5 w-5 text-white group-hover:translate-x-1 transition-all" />
        </div>
      </div>
    </button>
  );
}

