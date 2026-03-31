"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { X, ChevronLeft, ChevronRight, Thermometer, Droplets, Battery, ChevronRight as ChevronRightIcon, Loader2, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { sensorService } from "@/services/api/sensorService";
import { deviceLocationService } from "@/services/api/deviceLocationService";
import type { Device } from "@/types/device";
import type { Reading } from "@/types/admin";

interface DeviceCarouselProps {
  devices: Device[];
  currentIndex: number;
  onClose: () => void;
  onNavigate: (index: number) => void;
  onDeviceSelect?: (device: Device) => void;
  onDeviceRemove?: () => void;
}

export function DeviceCarousel({ devices, currentIndex, onClose, onNavigate, onDeviceSelect, onDeviceRemove }: DeviceCarouselProps) {
  const router = useRouter();
  const currentDevice = devices[currentIndex];
  const [latestReading, setLatestReading] = useState<Reading | null>(null);
  const [isLoadingReading, setIsLoadingReading] = useState(true);
  const [isDeleting, setIsDeleting] = useState(false);

  const handlePrevious = () => {
    const newIndex = currentIndex > 0 ? currentIndex - 1 : devices.length - 1;
    onNavigate(newIndex);
    if (onDeviceSelect && devices[newIndex]) {
      onDeviceSelect(devices[newIndex]);
    }
  };

  const handleNext = () => {
    const newIndex = currentIndex < devices.length - 1 ? currentIndex + 1 : 0;
    onNavigate(newIndex);
    if (onDeviceSelect && devices[newIndex]) {
      onDeviceSelect(devices[newIndex]);
    }
  };

  // Load latest reading when device changes
  useEffect(() => {
    if (currentDevice) {
      loadLatestReading();
      if (onDeviceSelect) {
        onDeviceSelect(currentDevice);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentIndex, currentDevice?.id, currentDevice?.pi_id, currentDevice?.device_id]);

  const loadLatestReading = async () => {
    if (!currentDevice) return;
    
    try {
      setIsLoadingReading(true);
      const reading = await sensorService.getLatestDeviceReading(
        currentDevice.pi_id,
        currentDevice.device_id
      );
      setLatestReading(reading);
    } catch (err) {
      console.error("Error loading latest reading:", err);
      setLatestReading(null);
    } finally {
      setIsLoadingReading(false);
    }
  };

  const handleViewAnalytics = () => {
    router.push(`/user/sensors/${currentDevice.device_id}?pi_id=${encodeURIComponent(currentDevice.pi_id)}`);
  };

  const handleRemoveFromMap = async () => {
    if (!currentDevice) return;
    
    // Confirm deletion
    if (!confirm("Are you sure you want to remove this device from the map?")) {
      return;
    }

    try {
      setIsDeleting(true);
      await deviceLocationService.deleteLocation(currentDevice.device_id);
      
      // Notify parent to refresh devices
      if (onDeviceRemove) {
        onDeviceRemove();
      }
      
      // Close the carousel
      onClose();
    } catch (err) {
      console.error("Error removing device from map:", err);
      alert(err instanceof Error ? err.message : "Failed to remove device from map");
    } finally {
      setIsDeleting(false);
    }
  };

  if (!currentDevice) return null;

  return (
    <div className="w-full max-w-2xl rounded-t-2xl overflow-hidden bg-black/98 backdrop-blur-xl border border-white/10 border-b-0 shadow-[0_-8px_32px_rgba(0,0,0,0.5)]">
      <div className="w-12 h-1 rounded-full bg-white/20 mx-auto mt-3 mb-1" />
      <div className="flex items-center justify-between px-4 pb-2">
        <div className="flex items-center gap-2">
          {devices.length > 1 && (
            <>
              <Button
                onClick={handlePrevious}
                className="bg-white/5 hover:bg-orange-500/20 h-8 w-8 p-0 rounded-xl transition-all text-white/80 hover:text-white"
              >
                <ChevronLeft className="h-4 w-4" />
              </Button>
              <span className="text-xs text-white/50 font-light min-w-[3rem] text-center">
                {currentIndex + 1} / {devices.length}
              </span>
              <Button
                onClick={handleNext}
                className="bg-white/5 hover:bg-orange-500/20 h-8 w-8 p-0 rounded-xl transition-all text-white/80 hover:text-white"
              >
                <ChevronRight className="h-4 w-4" />
              </Button>
            </>
          )}
        </div>
        <Button
          onClick={onClose}
          className="h-8 w-8 p-0 rounded-xl bg-white/5 hover:bg-white/10 text-white/70 hover:text-white transition-all"
        >
          <X className="h-4 w-4" />
        </Button>
      </div>

      <div className="p-4 pt-0 max-h-[60vh] overflow-y-auto">
        <div className="mb-4">
          <p className="text-sm font-mono font-light text-white/90">{currentDevice.device_id}</p>
          <p className="text-xs text-white/40 font-light mt-0.5">Pi: {currentDevice.pi_id}</p>
        </div>

        {isLoadingReading ? (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-5 w-5 text-white/40 animate-spin" />
          </div>
        ) : latestReading ? (
          <div className="space-y-4">
            <div className="grid grid-cols-3 gap-2">
              {latestReading.payload.sensors.temperature && (
                <div className="flex flex-col items-center gap-1.5 p-3 rounded-xl bg-white/5 border border-white/5">
                  <Thermometer className="h-4 w-4 text-orange-400/80" />
                  <span className="text-sm font-light text-white/90">
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
                <div className="flex flex-col items-center gap-0.5 p-3 rounded-xl bg-white/5 border border-white/5">
                  <Droplets className="h-4 w-4 text-orange-400/80" />
                  {latestReading.fill_percentage != null ? (
                    <span className="text-sm font-light text-white/90">{latestReading.fill_percentage.toFixed(0)}% fill</span>
                  ) : latestReading.sap_depth_cm != null ? (
                    <span className="text-sm font-light text-white/90">{latestReading.sap_depth_cm.toFixed(0)} cm sap</span>
                  ) : currentDevice.height != null ? (
                    <span className="text-sm font-light text-white/90">{Math.max(0, currentDevice.height - latestReading.payload.sensors.level.value).toFixed(0)} cm sap</span>
                  ) : (
                    <span className="text-sm font-light text-white/90">{latestReading.payload.sensors.level.value.toFixed(1)} cm to surface</span>
                  )}
                  {currentDevice.height != null && (
                    <span className="text-[10px] text-white/40 font-light">
                      {Math.max(0, currentDevice.height - latestReading.payload.sensors.level.value).toFixed(0)}cm sap · {latestReading.payload.sensors.level.value.toFixed(0)}cm to top
                    </span>
                  )}
                </div>
              )}
              <div className="flex flex-col items-center gap-1.5 p-3 rounded-xl bg-white/5 border border-white/5">
                <Battery className="h-4 w-4 text-orange-400/80" />
                <span className="text-sm font-light text-white/90">
                  {latestReading.payload.battery_percentage.toFixed(0)}%
                </span>
              </div>
            </div>
            <p className="text-xs text-white/40 font-light">
              Updated {new Date(latestReading.ts).toLocaleString()}
            </p>
          </div>
        ) : (
          <p className="text-sm text-white/40 font-light py-4">No readings yet</p>
        )}

        <div className="flex gap-2 mt-4 pt-4 border-t border-white/5">
          <Button
            onClick={handleViewAnalytics}
            className="flex-1 h-11 rounded-xl bg-orange-500/90 hover:bg-orange-500 text-white font-light transition-all flex items-center justify-center gap-2"
          >
            <span className="text-sm font-light">View Analytics</span>
            <ChevronRightIcon className="h-4 w-4" />
          </Button>
          <Button
            onClick={handleRemoveFromMap}
            disabled={isDeleting}
            className="h-11 rounded-xl bg-white/5 hover:bg-red-500/20 border border-white/10 text-white/80 hover:text-red-400 transition-colors flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed px-4"
          >
            {isDeleting ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Trash2 className="h-4 w-4" />
            )}
          </Button>
        </div>
      </div>
    </div>
  );
}

