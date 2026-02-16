"use client";

import { useState, useEffect } from "react";
import { useRequireAuth } from "@/hooks/useRequireAuth";
import Navbar from "@/components/navbar/Navbar";
import { LocationInput } from "@/components/weather/LocationInput";
import { CurrentWeather } from "@/components/weather/CurrentWeather";
import { usePiPreferences } from "@/hooks/usePiPreferences";
import dynamic from "next/dynamic";

const DeviceMap = dynamic(
  () =>
    import("@/components/map/DeviceMap")
      .then((mod) => ({ default: mod.DeviceMap }))
      .catch((error) => {
        console.error("Failed to load DeviceMap:", error);
        return {
          default: function DeviceMapFallback() {
            return (
              <div className="w-full h-[600px] rounded-lg border border-white/10 flex items-center justify-center bg-black/50">
                <div className="text-white/60 font-light">Failed to load map. Please refresh the page.</div>
              </div>
            );
          },
        };
      }),
  {
    ssr: false,
    loading: () => (
      <div className="w-full h-[600px] rounded-lg border border-white/10 flex items-center justify-center bg-black/50">
        <div className="text-white/60 font-light">Loading map...</div>
      </div>
    ),
  }
);
import { DeviceCarousel } from "@/components/map/DeviceCarousel";
import { weatherService } from "@/services/api/weatherService";
import { deviceLocationService } from "@/services/api/deviceLocationService";
import { sensorService, type Pi, type PiDevice } from "@/services/api/sensorService";
import { ErrorBoundary } from "@/components/ErrorBoundary";
import type { CurrentWeather as CurrentWeatherType } from "@/types/weather";
import type { Device } from "@/types/device";
import { Loader2, AlertCircle } from "lucide-react";

export default function UserDashboardPage() {
  const { user, isLoading } = useRequireAuth("user");
  const { getPreference } = usePiPreferences();
  const [hasLocation, setHasLocation] = useState<boolean | null>(null);
  const [isCheckingLocation, setIsCheckingLocation] = useState(true);
  const [isUpdatingLocation, setIsUpdatingLocation] = useState(false);
  const [currentWeather, setCurrentWeather] = useState<CurrentWeatherType | null>(null);
  const [isLoadingWeather, setIsLoadingWeather] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [locationName, setLocationName] = useState<string>("");
  const [latitude, setLatitude] = useState<number | null>(null);
  const [longitude, setLongitude] = useState<number | null>(null);
  const [showLocationInput, setShowLocationInput] = useState(false);
  
  // Device management
  const [devices, setDevices] = useState<Device[]>([]);
  const [selectedDeviceIndex, setSelectedDeviceIndex] = useState<number | null>(null);
  const [selectedDeviceId, setSelectedDeviceId] = useState<string | null>(null);
  const [showCarousel, setShowCarousel] = useState(false);
  
  // Available PIs and devices for dropdowns
  const [availablePis, setAvailablePis] = useState<Pi[]>([]);
  const [availableDevices, setAvailableDevices] = useState<PiDevice[]>([]);
  const [isLoadingSensors, setIsLoadingSensors] = useState(false);

  // Check if user has location set and load devices
  useEffect(() => {
    if (!isLoading && user) {
      checkLocation();
      loadDevices();
      loadAvailableSensors();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isLoading, user]);

  // Fetch weather data when location is available
  useEffect(() => {
    if (hasLocation === true) {
      fetchWeatherData();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [hasLocation]);

  const checkLocation = async () => {
    try {
      setIsCheckingLocation(true);
      const profile = await weatherService.getProfile();
      if (profile && profile.latitude !== null && profile.latitude !== undefined && 
          profile.longitude !== null && profile.longitude !== undefined) {
        setHasLocation(true);
        setLatitude(profile.latitude);
        setLongitude(profile.longitude);
      } else {
        // User hasn't set location yet - this is normal for new users
        setHasLocation(false);
        setShowLocationInput(true);
        setError(null); // Clear any previous errors
      }
    } catch (err) {
      console.error("Error checking location:", err);
      // If it's a token refresh error or 401/404, treat as "no location set"
      const errorMessage = err instanceof Error ? err.message : "";
      if (errorMessage.includes("Token refresh failed") || 
          errorMessage.includes("401") || 
          errorMessage.includes("404")) {
        setHasLocation(false);
        setShowLocationInput(true);
        setError(null); // Don't show error for new users without location
      } else {
        // Only show error for unexpected errors
        setError("Failed to check location status");
        setHasLocation(false);
        setShowLocationInput(true);
      }
    } finally {
      setIsCheckingLocation(false);
    }
  };

  const handleLocationSubmit = async (lat: number, lon: number, name: string) => {
    try {
      setIsUpdatingLocation(true);
      setError(null);
      await weatherService.updateLocation({ latitude: lat, longitude: lon });
      setLocationName(name);
      setLatitude(lat);
      setLongitude(lon);
      setHasLocation(true);
      setShowLocationInput(false);
    } catch (err) {
      console.error("Error updating location:", err);
      const errorMessage = err instanceof Error ? err.message : "Failed to update location";
      setError(errorMessage);
    } finally {
      setIsUpdatingLocation(false);
    }
  };

  const fetchWeatherData = async () => {
    try {
      setIsLoadingWeather(true);
      setError(null);
      const current = await weatherService.getCurrentWeather();
      setCurrentWeather(current);
      if (!locationName && current.name) {
        setLocationName(current.name);
      }
      if (current.coord) {
        setLatitude(current.coord.lat);
        setLongitude(current.coord.lon);
      }
    } catch (err) {
      console.error("Error fetching weather:", err);
      const errorMessage = err instanceof Error ? err.message : "Failed to fetch weather data";
      if (errorMessage.includes("location not set") || 
          errorMessage.includes("Token refresh failed") ||
          errorMessage.includes("401") ||
          errorMessage.includes("404")) {
        // User hasn't set location - show friendly message instead of error
        setHasLocation(false);
        setError(null); // Clear error - we'll show location input instead
        setShowLocationInput(true);
      } else {
        setError(errorMessage);
      }
    } finally {
      setIsLoadingWeather(false);
    }
  };

  const loadDevices = async () => {
    try {
      const locations = await deviceLocationService.getAllLocations();
      const convertedDevices = locations.map((location) =>
        deviceLocationService.convertToDevice(location, location.device_id)
      );
      setDevices(convertedDevices);
      
      // Load colors from backend after devices are loaded
      // This will sync colors from device locations to PI preferences
      // Note: This is handled by the DeviceMap component's usePiPreferences hook
    } catch (err) {
      console.error("Error loading devices:", err);
      const errorMessage = err instanceof Error ? err.message : "Failed to load devices";
      // Only show error if it's a real error that needs user attention
      // Network errors and empty states are handled gracefully by the service
      if (errorMessage && !errorMessage.toLowerCase().includes("network") && !errorMessage.includes("404")) {
        // Only show non-critical errors
        console.warn("Device loading warning:", errorMessage);
      }
      // Set empty array on error - user can still add devices
      setDevices([]);
    }
  };

  const loadAvailableSensors = async () => {
    if (!user?.user_id) return;

    try {
      setIsLoadingSensors(true);
      
      // Fetch all PIs for the user
      const pisResponse = await sensorService.getPis({
        user_id: user.user_id,
        page: 1,
        page_size: 100, // Get all PIs
      });

      // Handle null/undefined items - new users may not have PIs yet
      const pis = Array.isArray(pisResponse?.items) ? pisResponse.items : [];
      setAvailablePis(pis);

      // Fetch devices for each PI
      const allDevices: PiDevice[] = [];
      if (pis.length > 0) {
        await Promise.all(
          pis.map(async (pi) => {
            try {
              const devicesResponse = await sensorService.getDevices({
                pi_id: pi.pi_id,
                page: 1,
                page_size: 100, // Get all devices
              });
              // Handle null/undefined items
              const devices = Array.isArray(devicesResponse?.items) ? devicesResponse.items : [];
              allDevices.push(...devices);
            } catch (err) {
              console.error(`Error loading devices for PI ${pi.pi_id}:`, err);
            }
          })
        );
      }

      setAvailableDevices(allDevices);
    } catch (err) {
      console.error("Error loading available sensors:", err);
      // Don't show error to user - they can still manually type if needed
      setAvailablePis([]);
      setAvailableDevices([]);
    } finally {
      setIsLoadingSensors(false);
    }
  };

  const handleDeviceAdd = async (deviceData: Omit<Device, "id" | "createdAt" | "updatedAt">) => {
    try {
      setError(null);
      // Get the gradient preference for this PI (with default fallback)
      const piPreference = getPreference(deviceData.pi_id);
      // Get color from preference (with default fallback)
      const hexColor = piPreference?.color || "#f97316";
      
      const location = await deviceLocationService.addLocation({
        device_id: deviceData.device_id,
        pi_id: deviceData.pi_id,
        latitude: deviceData.latitude,
        longitude: deviceData.longitude,
        color: hexColor,
      });
      // Reload all devices to ensure consistency
      await loadDevices();
    } catch (err) {
      console.error("Error adding device:", err);
      const errorMessage = err instanceof Error ? err.message : "Failed to add device";
      setError(errorMessage);
      throw err; // Re-throw so the form can handle it
    }
  };

  const handleDeviceClick = (device: Device) => {
    const index = devices.findIndex((d) => d.id === device.id);
    if (index !== -1) {
      setSelectedDeviceIndex(index);
      setSelectedDeviceId(device.id);
      setShowCarousel(true);
    }
  };

  const handleDeviceSelect = (device: Device) => {
    setSelectedDeviceId(device.id);
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

  const mapCenter: [number, number] = latitude && longitude 
    ? [latitude, longitude] 
    : [40.7580, -74.0390]; // Default to Weehawken, NJ

  return (
    <div className="min-h-screen bg-black text-white">
      <Navbar />
      <main className="pt-24 px-4 sm:px-6 lg:px-8 pb-16">
        <div className="max-w-7xl mx-auto">
          <div className="mb-8">
            <h1 className="text-4xl font-light tracking-tight mb-2">
              Dashboard
            </h1>
            {/* <p className="text-white/60 font-light text-sm">
              Welcome back, {user?.username}
            </p> */}
          </div>

          {error && (
            <div className="mb-6 p-4 border border-red-500/20 bg-red-500/10 rounded-lg flex items-center gap-2">
              <AlertCircle className="h-5 w-5 text-red-400" />
              <p className="text-sm text-red-400 font-light">{error}</p>
            </div>
          )}

          {/* Location Input Panel - Only show if location not set */}
          {showLocationInput && !hasLocation && (
            <div className="border border-white/10 rounded-lg p-8 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm mb-6">
              <h2 className="text-2xl font-light mb-4">Welcome! Set Your Location</h2>
              <p className="text-orange-400/80 font-light text-sm mb-6">
                To get started, please set your location to view weather information and manage your sensors.
              </p>
              <LocationInput
                onLocationSubmit={handleLocationSubmit}
                isLoading={isUpdatingLocation}
              />
            </div>
          )}

          {/* Weather Panel */}
          {hasLocation && (
            <div className="mb-6 border border-white/10 rounded-lg bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm overflow-hidden">
              <div className="px-4 pb-4 pt-4">
                {isLoadingWeather ? (
                  <div className="flex items-center justify-center py-8">
                    <div className="flex flex-col items-center gap-2">
                      <Loader2 className="h-6 w-6 text-white/60 animate-spin" />
                      <p className="text-white/60 font-light text-sm">Loading weather data...</p>
                    </div>
                  </div>
                ) : currentWeather ? (
                  <CurrentWeather weather={currentWeather} />
                ) : (
                  <p className="text-white/60 font-light text-center py-6 text-sm">
                    No weather data available
                  </p>
                )}
              </div>
            </div>
          )}

          {/* Interactive Map */}
          <div className="mb-6 border border-white/10 rounded-lg p-6 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm">
            <h2 className="text-xl font-light mb-4">Sensor Map</h2>
            <p className="text-white/60 font-light text-sm mb-4">
              Click &quot;Add Device&quot; to place sensors on the map. Click on markers to view device information.
            </p>
            <ErrorBoundary
              fallback={
                <div className="w-full h-[600px] rounded-lg border border-red-500/20 bg-red-500/10 flex items-center justify-center">
                  <div className="text-center">
                    <p className="text-red-400 font-light mb-2">Failed to load map</p>
                    <button
                      onClick={() => window.location.reload()}
                      className="text-white/60 hover:text-white text-sm underline"
                    >
                      Reload page
                    </button>
                  </div>
                </div>
              }
            >
              <DeviceMap
                devices={devices}
                onDeviceAdd={handleDeviceAdd}
                onDeviceClick={handleDeviceClick}
                center={mapCenter}
                availablePis={availablePis}
                availableDevices={availableDevices}
                selectedDeviceId={selectedDeviceId}
                carousel={
                  showCarousel && selectedDeviceIndex !== null ? (
                    <DeviceCarousel
                      devices={devices}
                      currentIndex={selectedDeviceIndex}
                      onClose={() => {
                        setShowCarousel(false);
                        setSelectedDeviceIndex(null);
                        setSelectedDeviceId(null);
                      }}
                      onNavigate={(index) => {
                        setSelectedDeviceIndex(index);
                        if (devices[index]) {
                          setSelectedDeviceId(devices[index].id);
                        }
                      }}
                      onDeviceSelect={handleDeviceSelect}
                      onDeviceRemove={loadDevices}
                    />
                  ) : null
                }
              />
            </ErrorBoundary>
          </div>
        </div>
      </main>

    </div>
  );
}
