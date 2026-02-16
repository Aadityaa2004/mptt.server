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
    <div className="max-w-md w-full">
      <div className="relative border border-white/20 rounded-lg bg-black/40 backdrop-blur-lg shadow-xl">
        {/* Close Button */}
        <Button
          onClick={onClose}
          className="absolute top-2 right-2 z-10 bg-white/10 hover:bg-orange-500/20 border border-white/20 hover:border-orange-500/40 text-white hover:text-orange-200 h-8 w-8 p-0 transition-all"
        >
          <X className="h-4 w-4" />
        </Button>

        {/* Device Info */}
        <div className="p-6 max-h-[70vh] overflow-y-auto">
          <div className="mb-4">
            <h2 className="text-xl font-light mb-1">Device</h2>
            <p className="text-white/60 font-light text-xs font-mono tracking-wider mb-2">
              {currentDevice.device_id}
            </p>
            <p className="text-xs text-white/40 font-light">Pi ID: {currentDevice.pi_id}</p>
          </div>

          {/* Latest Reading */}
          {isLoadingReading ? (
            <div className="mt-4 pt-4 border-t border-white/10 flex items-center justify-center py-4">
              <Loader2 className="h-4 w-4 text-white/40 animate-spin" />
            </div>
          ) : latestReading ? (
            <div className="mt-4 pt-4 border-t border-white/10 space-y-3">
              {/* Sensor Readings Grid - Similar to DeviceCard */}
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

          {/* Action Buttons */}
          <div className="mt-4 pt-4 border-t border-white/10 flex gap-2">
            <Button
              onClick={handleRemoveFromMap}
              disabled={isDeleting}
              className="flex-1 bg-red-500/20 hover:bg-red-500/30 border border-red-500/30 text-red-400 hover:text-red-300 transition-colors flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isDeleting ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  <span className="text-sm font-light">Removing...</span>
                </>
              ) : (
                <>
                  <Trash2 className="h-4 w-4" />
                  <span className="text-sm font-light">Remove from Map</span>
                </>
              )}
            </Button>
            <Button
              onClick={handleViewAnalytics}
              className="flex-1 bg-gray-300 hover:bg-gray-300/90 border border-white/10 text-black transition-colors flex items-center justify-center gap-2"
            >
              <span className="text-sm font-light">View Analytics</span>
              <ChevronRightIcon className="h-4 w-4" />
            </Button>
          </div>
        </div>

        {/* Navigation */}
        {devices.length > 1 && (
          <div className="flex items-center justify-between px-4 pb-4 pt-2 border-t border-white/10">
            <Button
              onClick={handlePrevious}
              className="bg-white/10 hover:bg-orange-500/20 border border-white/20 hover:border-orange-500/40 text-white hover:text-orange-200 h-8 w-8 p-0 rounded-full transition-all"
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <div className="text-xs text-white/60 font-light bg-white/5 px-3 py-1 rounded-full border border-white/10">
              {currentIndex + 1} / {devices.length}
            </div>
            <Button
              onClick={handleNext}
              className="bg-white/10 hover:bg-orange-500/20 border border-white/20 hover:border-orange-500/40 text-white hover:text-orange-200 h-8 w-8 p-0 rounded-full transition-all"
            >
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}

