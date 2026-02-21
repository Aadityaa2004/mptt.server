"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { sensorService } from "@/services/api/sensorService";
import type { Reading } from "@/types/admin";
import { ChevronRight, Thermometer, Droplets, Battery, Loader2 } from "lucide-react";

interface DeviceCardProps {
  device: {
    pi_id: string;
    device_id: string;
  };
}

export function DeviceCard({ device }: DeviceCardProps) {
  const router = useRouter();
  const [latestReading, setLatestReading] = useState<Reading | null>(null);
  const [deviceDims, setDeviceDims] = useState<{ height?: number } | null>(null);
  const [isLoadingReading, setIsLoadingReading] = useState(true);
  const [collectionEnabled, setCollectionEnabled] = useState<boolean | null>(null);
  const [isTogglingCollection, setIsTogglingCollection] = useState(false);

  useEffect(() => {
    loadLatestReading();
    sensorService.getDevice(device.pi_id, device.device_id).then((d) => {
      setCollectionEnabled(d.collection_enabled !== false);
      if (d.height != null) setDeviceDims({ height: d.height });
    }).catch(() => setCollectionEnabled(null));
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
    router.push(`/user/sensors/${device.device_id}?pi_id=${encodeURIComponent(device.pi_id)}`);
  };

  const handleToggleCollection = async (e: React.MouseEvent) => {
    e.stopPropagation();
    if (collectionEnabled === null || isTogglingCollection) return;
    try {
      setIsTogglingCollection(true);
      await sensorService.updateDevice(device.pi_id, device.device_id, { collection_enabled: !collectionEnabled });
      setCollectionEnabled(!collectionEnabled);
    } finally {
      setIsTogglingCollection(false);
    }
  };

  return (
    <button
      onClick={handleClick}
      className="group flex flex-col p-5 rounded-xl bg-white/[0.04] border border-white/5 hover:border-white/10 hover:bg-white/[0.06] transition-all text-left w-full h-full"
    >
      <div className="flex items-start justify-between mb-3">
        <p className="text-sm font-mono font-light text-white/90 truncate" title={device.device_id}>
          {device.device_id}
        </p>
        {collectionEnabled !== null && (
          <span
            onClick={(e) => e.stopPropagation()}
            className="flex-shrink-0"
          >
            <button
              type="button"
              onClick={handleToggleCollection}
              disabled={isTogglingCollection}
              className={`px-2.5 py-1 rounded-lg text-xs font-light transition-colors ${collectionEnabled ? "bg-green-500/20 text-green-400" : "bg-white/5 text-white/50"}`}
            >
              {isTogglingCollection ? "..." : collectionEnabled ? "On" : "Off"}
            </button>
          </span>
        )}
      </div>

      {isLoadingReading ? (
        <div className="flex items-center justify-center py-6">
          <Loader2 className="h-4 w-4 text-white/40 animate-spin" />
        </div>
      ) : latestReading ? (
        <div className="space-y-3">
          <div className="grid grid-cols-3 gap-2">
            {latestReading.payload.sensors.temperature && (
              <div className="flex flex-col items-center gap-1 p-2.5 rounded-lg bg-white/5">
                <Thermometer className="h-3.5 w-3.5 text-orange-400/80" />
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
              <div className="flex flex-col items-center gap-0.5 p-2.5 rounded-lg bg-white/5">
                <Droplets className="h-3.5 w-3.5 text-orange-400/80" />
                {latestReading.fill_percentage != null ? (
                  <span className="text-xs font-light text-white/90">{latestReading.fill_percentage.toFixed(0)}% fill</span>
                ) : (
                  <>
                    <span className="text-xs font-light text-white/90">{latestReading.payload.sensors.level.value.toFixed(1)}</span>
                    <span className="text-[10px] font-light text-white/50">{latestReading.payload.sensors.level.unit}</span>
                  </>
                )}
                {deviceDims?.height != null && (
                  <span className="text-[9px] text-white/40 font-light">
                    {Math.max(0, deviceDims.height - latestReading.payload.sensors.level.value).toFixed(0)}cm sap · {latestReading.payload.sensors.level.value.toFixed(0)}cm left
                  </span>
                )}
              </div>
            )}
            <div className="flex flex-col items-center gap-1 p-2.5 rounded-lg bg-white/5">
              <Battery className="h-3.5 w-3.5 text-orange-400/80" />
              <span className="text-xs font-light text-white/90">
                {latestReading.payload.battery_percentage.toFixed(0)}%
              </span>
            </div>
          </div>
          <p className="text-[11px] text-white/40 font-light">
            {new Date(latestReading.ts).toLocaleString()}
          </p>
        </div>
      ) : (
        <p className="text-xs text-white/45 font-light py-4">No readings yet</p>
      )}

      <div className="mt-4 pt-3 border-t border-white/5 flex items-center justify-end">
        <span className="inline-flex items-center gap-1 text-xs font-light text-white/60 group-hover:text-orange-400 transition-colors">
          View
          <ChevronRight className="h-4 w-4 group-hover:translate-x-0.5 transition-transform" />
        </span>
      </div>
    </button>
  );
}

