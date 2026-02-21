"use client";

import { useState, useEffect } from "react";
import { useRouter, useParams, useSearchParams } from "next/navigation";
import { useRequireAuth } from "@/hooks/useRequireAuth";
import Navbar from "@/components/navbar/Navbar";
import { sensorService } from "@/services/api/sensorService";
import { ReadingsTable } from "@/components/sensors/ReadingsTable";
import { ReadingsChart } from "@/components/sensors/ReadingsChart";
import type { Reading } from "@/types/admin";
import { Loader2, AlertCircle, ArrowLeft, Thermometer, Droplets, Battery, Copy, Check, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";

export default function SensorAnalyticsPage() {
  const { user, isLoading } = useRequireAuth("user");
  const router = useRouter();
  const params = useParams();
  const searchParams = useSearchParams();
  // Decode device_id in case it's URL-encoded
  const deviceId = params?.device_id ? decodeURIComponent(params.device_id as string) : "";
  const piIdFromQuery = searchParams?.get("pi_id");

  const [piId, setPiId] = useState<string | null>(null);
  const [latestReading, setLatestReading] = useState<Reading | null>(null);
  const [readings, setReadings] = useState<Reading[]>([]);
  const [isLoadingReadings, setIsLoadingReadings] = useState(true);
  const [isFindingPi, setIsFindingPi] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [timeRange, setTimeRange] = useState<"1h" | "1d" | "1w" | "1m" | "1y">("1d");
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [collectionEnabled, setCollectionEnabled] = useState<boolean | null>(null);
  const [isTogglingCollection, setIsTogglingCollection] = useState(false);
  const [deviceHeight, setDeviceHeight] = useState<number | null>(null);

  useEffect(() => {
    if (!isLoading && user && deviceId) {
      // If pi_id is in query params, use it directly
      if (piIdFromQuery) {
        setPiId(piIdFromQuery);
        setIsFindingPi(false);
      } else if (!piId) {
        // Only try to find it if we don't already have a piId
        findPiForDevice();
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isLoading, user, deviceId]);
  
  // Separate effect to handle piIdFromQuery changes
  useEffect(() => {
    if (piIdFromQuery && piIdFromQuery !== piId) {
      setPiId(piIdFromQuery);
      setIsFindingPi(false);
    }
  }, [piIdFromQuery, piId]);

  useEffect(() => {
    if (piId && deviceId) {
      loadLatestReading();
      loadReadingsForStats();
      sensorService.getDevice(piId, deviceId).then((d) => {
        setCollectionEnabled(d.collection_enabled !== false);
        if (d.height != null) setDeviceHeight(d.height);
      }).catch(() => setCollectionEnabled(null));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [piId, deviceId]);

  const findPiForDevice = async () => {
    if (!user?.user_id) return;

    try {
      setIsFindingPi(true);
      setError(null);
      
      console.log("Finding PI for device:", deviceId);
      
      // Fetch all PIs for the user
      const pisResponse = await sensorService.getPis({
        user_id: user.user_id,
        page: 1,
        page_size: 100,
      });

      console.log("Found PIs:", pisResponse.items.length);

      // Try to fetch readings for each PI with this device_id
      // This is the most reliable way since the readings endpoint works
      for (const pi of pisResponse.items) {
        try {
          console.log(`Trying PI ${pi.pi_id} for device ${deviceId}`);
          const response = await sensorService.getDeviceReadings(pi.pi_id, deviceId, {
            limit: 1,
          });
          
          // If we got a response (even with empty items), this PI has the device
          console.log(`Found readings for PI ${pi.pi_id}:`, response.items.length);
          setPiId(pi.pi_id);
          setIsFindingPi(false);
          return;
        } catch (err: any) {
          // Check if it's a 404 - that means device doesn't exist on this PI
          // Other errors might be network issues, so we should continue
          const is404 = err?.message?.includes("404") || err?.message?.includes("not found");
          if (is404) {
            console.log(`PI ${pi.pi_id} doesn't have this device (404), trying next...`);
            continue;
          } else {
            // For other errors, log but continue
            console.error(`Error checking PI ${pi.pi_id}:`, err);
            continue;
          }
        }
      }

      // Fallback: Try the device list approach
      console.log("Fallback: checking device lists...");
      for (const pi of pisResponse.items) {
        try {
          const devicesResponse = await sensorService.getDevices({
            pi_id: pi.pi_id,
            page: 1,
            page_size: 100,
          });

          // Try both exact match and case-insensitive match
          const device = devicesResponse.items.find((d) => 
            d.device_id === deviceId || 
            d.device_id?.toUpperCase() === deviceId.toUpperCase() ||
            decodeURIComponent(d.device_id || "") === deviceId
          );
          
          if (device) {
            console.log(`Found device in PI ${pi.pi_id} via device list`);
            setPiId(pi.pi_id);
            setIsFindingPi(false);
            return;
          }
        } catch (err) {
          console.error(`Error checking devices for PI ${pi.pi_id}:`, err);
        }
      }

      console.error("Device not found in any PI. Device ID:", deviceId);
      console.error("Available PIs:", pisResponse.items.map(p => p.pi_id));
      setError(`Device "${deviceId}" not found or you don't have access to it`);
      setIsFindingPi(false);
    } catch (err) {
      console.error("Error finding PI for device:", err);
      const errorMessage = err instanceof Error ? err.message : "Failed to find device";
      setError(errorMessage);
      setIsFindingPi(false);
    }
  };

  const loadLatestReading = async () => {
    if (!piId || !deviceId) return;

    try {
      // Use the getReadings endpoint with limit=1 to get the latest reading
      const response = await sensorService.getReadings(piId, deviceId, {
        limit: 1,
      });
      
      if (response.items.length > 0) {
        setLatestReading(response.items[0]);
      }
    } catch (err) {
      console.error("Error loading latest reading:", err);
      // Don't set error for latest reading failures, just log
    }
  };

  const loadReadingsForStats = async () => {
    if (!piId || !deviceId) return;

    try {
      setIsLoadingReadings(true);
      setError(null);

      // Use deviceId as-is (MAC address string)
      const response = await sensorService.getDeviceReadings(piId, deviceId, {
        page: 1,
        limit: 100, // Get more readings for better stats
      });

      // Ensure readings is always an array
      setReadings(Array.isArray(response?.items) ? response.items : []);
    } catch (err) {
      console.error("Error loading readings for stats:", err);
      // Set to empty array on error to prevent undefined/null issues
      setReadings([]);
      // Don't set error here, just log - stats are optional
    } finally {
      setIsLoadingReadings(false);
    }
  };

  const formatDateShort = (timestamp: string) => {
    return new Date(timestamp).toLocaleString("en-US", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const copyDeviceId = async () => {
    try {
      await navigator.clipboard.writeText(deviceId);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy device ID:", err);
    }
  };

  const refreshSensorData = async () => {
    if (!piId || !deviceId || isRefreshing) return;

    try {
      setIsRefreshing(true);
      setError(null);
      
      // Refresh both latest reading and readings for stats
      await Promise.all([
        loadLatestReading(),
        loadReadingsForStats()
      ]);
    } catch (err) {
      console.error("Error refreshing sensor data:", err);
      const errorMessage = err instanceof Error ? err.message : "Failed to refresh sensor data";
      setError(errorMessage);
    } finally {
      setIsRefreshing(false);
    }
  };

  const handleToggleCollection = async () => {
    if (!piId || !deviceId || collectionEnabled === null || isTogglingCollection) return;
    try {
      setIsTogglingCollection(true);
      setError(null);
      await sensorService.updateDevice(piId, deviceId, { collection_enabled: !collectionEnabled });
      setCollectionEnabled(!collectionEnabled);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update collection");
    } finally {
      setIsTogglingCollection(false);
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-background text-foreground flex items-center justify-center">
        <div className="flex flex-col items-center gap-3">
          <Loader2 className="h-6 w-6 text-white/60 animate-spin" />
          <p className="text-white/60 font-light">Loading...</p>
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
          <div className="px-4 sm:px-6 lg:px-8 py-6 border-b border-white/5">
            <Button
              variant="ghost"
              onClick={() => router.push("/user/sensors")}
              className="mb-4 -ml-2 text-white/70 hover:text-white hover:bg-white/5 px-3 py-2 rounded-xl font-light text-sm"
            >
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back to Sensors
            </Button>
            <div className="flex flex-wrap items-center justify-between gap-4">
              <div>
                <h1 className="text-2xl sm:text-3xl font-light tracking-tight text-foreground">
                  Sensor Analytics
                </h1>
                <div className="flex items-center gap-2 mt-2">
                  <p className="text-white/60 font-light text-sm font-mono">
                    {deviceId}
                  </p>
                  <button
                    onClick={copyDeviceId}
                    className="p-1.5 rounded-lg hover:bg-white/10 text-white/60 hover:text-white transition-colors"
                    title="Copy"
                  >
                    {copied ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
                  </button>
                </div>
                {piId && collectionEnabled !== null && (
                  <div className="flex items-center gap-2 mt-2">
                    <span className="text-xs text-white/50 font-light">Collection</span>
                    <button
                      onClick={handleToggleCollection}
                      disabled={isTogglingCollection}
                      className={`px-2.5 py-1 rounded-lg text-xs font-light transition-colors ${collectionEnabled ? "bg-green-500/20 text-green-400" : "bg-white/5 text-white/50"}`}
                    >
                      {isTogglingCollection ? "..." : collectionEnabled ? "On" : "Off"}
                    </button>
                  </div>
                )}
              </div>
            </div>
          </div>

          {error && (
            <div className="mx-4 sm:mx-6 lg:mx-8 mt-4 p-4 rounded-xl bg-red-500/5 border border-red-500/20 flex items-center gap-2">
              <AlertCircle className="h-5 w-5 text-red-400 flex-shrink-0" />
              <p className="text-sm text-red-400 font-light">{error}</p>
            </div>
          )}

          {/* Current Reading */}
          {latestReading && (
            <section className="px-4 sm:px-6 lg:px-8 py-6 border-b border-white/5">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-light text-white/90">Current Reading</h2>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={refreshSensorData}
                  disabled={isRefreshing}
                  className="h-9 w-9 rounded-xl bg-white/5 hover:bg-orange-500/20 text-white/70 hover:text-orange-400 disabled:opacity-50"
                  title="Refresh"
                >
                  <RefreshCw className={`h-4 w-4 ${isRefreshing ? "animate-spin" : ""}`} />
                </Button>
              </div>
              <div className="flex flex-wrap gap-3 sm:gap-4">
                {latestReading.payload.sensors.temperature && (
                  <div className="flex items-center gap-4 p-4 rounded-xl bg-white/[0.04] border border-white/5 min-w-[140px]">
                    <Thermometer className="h-5 w-5 text-orange-400/80" />
                    <div>
                      <div className="text-xl font-light">
                        {latestReading.payload.sensors.temperature.value.toFixed(1)}°
                        {latestReading.payload.sensors.temperature.unit === "fahrenheit" || latestReading.payload.sensors.temperature.unit === "F" ? "F" : latestReading.payload.sensors.temperature.unit === "celsius" || latestReading.payload.sensors.temperature.unit === "C" ? "C" : latestReading.payload.sensors.temperature.unit?.toUpperCase()}
                      </div>
                      <p className="text-xs text-white/50 font-light">Temperature</p>
                    </div>
                  </div>
                )}
                {latestReading.payload.sensors.level && (
                  <div className="flex items-center gap-4 p-4 rounded-xl bg-white/[0.04] border border-white/5 min-w-[140px]">
                    <Droplets className="h-5 w-5 text-orange-400/80" />
                    <div>
                      {latestReading.fill_percentage != null ? (
                        <>
                          <div className="text-xl font-light">{latestReading.fill_percentage.toFixed(1)}% fill</div>
                          <p className="text-xs text-white/50 font-light">Sap Level</p>
                          {deviceHeight != null && (
                            <p className="text-[11px] text-white/40 font-light mt-0.5">
                              {Math.max(0, deviceHeight - latestReading.payload.sensors.level.value).toFixed(0)}cm sap · {latestReading.payload.sensors.level.value.toFixed(0)}cm remaining
                            </p>
                          )}
                        </>
                      ) : (
                        <>
                          <div className="text-xl font-light">
                            {latestReading.payload.sensors.level.value.toFixed(1)} {latestReading.payload.sensors.level.unit}
                          </div>
                          <p className="text-xs text-white/50 font-light">Sap Level</p>
                          {deviceHeight != null && (
                            <p className="text-[11px] text-white/40 font-light mt-0.5">
                              {Math.max(0, deviceHeight - latestReading.payload.sensors.level.value).toFixed(0)}cm sap · {latestReading.payload.sensors.level.value.toFixed(0)}cm remaining
                            </p>
                          )}
                        </>
                      )}
                    </div>
                  </div>
                )}
                <div className="flex items-center gap-4 p-4 rounded-xl bg-white/[0.04] border border-white/5 min-w-[140px]">
                  <Battery className="h-5 w-5 text-orange-400/80" />
                  <div>
                    <div className="text-xl font-light">{latestReading.payload.battery_percentage.toFixed(0)}%</div>
                    <p className="text-xs text-white/50 font-light">Battery</p>
                  </div>
                </div>
                <div className="flex items-center self-center text-xs text-white/45 font-light ml-auto">
                  {formatDateShort(latestReading.ts)}
                </div>
              </div>
            </section>
          )}


          {/* Readings Chart */}
          {readings && Array.isArray(readings) && readings.length > 0 && (
            <section className="px-4 sm:px-6 lg:px-8 py-6 border-b border-white/5">
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-4">
                <h2 className="text-lg font-light text-white/90">Readings History</h2>
                <div className="flex flex-wrap gap-2">
                  {(["1h", "1d", "1w", "1m", "1y"] as const).map((range) => (
                    <button
                      key={range}
                      onClick={() => setTimeRange(range)}
                      className={`px-4 py-2 rounded-xl text-sm font-light transition-all ${
                        timeRange === range
                          ? "bg-orange-500/90 text-white"
                          : "bg-white/5 text-white/70 hover:bg-white/10 hover:text-white"
                      }`}
                    >
                      {range === "1h" ? "1H" : range === "1d" ? "1D" : range === "1w" ? "1W" : range === "1m" ? "1M" : "1Y"}
                    </button>
                  ))}
                </div>
              </div>
              <ReadingsChart readings={readings} timeRange={timeRange} />
            </section>
          )}

          {/* Readings Table */}
          <section className="px-4 sm:px-6 lg:px-8 py-6">
            <h2 className="text-lg font-light text-white/90 mb-4">Data Log</h2>
            <div className="rounded-2xl overflow-hidden border border-white/5 bg-white/[0.02]">
              {piId ? (
                <ReadingsTable deviceId={deviceId} piId={piId} />
              ) : isFindingPi ? (
                <div className="p-16 flex items-center justify-center">
                  <div className="flex flex-col items-center gap-3">
                    <Loader2 className="h-8 w-8 text-white/40 animate-spin" />
                    <p className="text-white/50 font-light text-sm">Loading readings...</p>
                  </div>
                </div>
              ) : (
                <div className="p-16 text-center">
                  <p className="text-white/50 font-light text-sm">Loading device...</p>
                </div>
              )}
            </div>
          </section>
        </div>
      </main>
    </div>
  );
}

