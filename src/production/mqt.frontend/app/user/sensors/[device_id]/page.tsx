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


  if (isLoading) {
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
        <div className="max-w-7xl mx-auto">
          {/* Header */}
          <div className="mb-8">
            <Button
              variant="ghost"
              onClick={() => router.push("/user/sensors")}
              className="mb-5 text-black bg-white/80 hover:bg-white/70 hover:text-black"
            >
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back to Sensors
            </Button>
            <h1 className="text-4xl font-light tracking-tight mb-2 text-white">
              Sensor Analytics
            </h1>
            <div className="flex items-center gap-2">
              <p className="text-white/60 font-light text-sm font-mono">
                Device ID: {deviceId}
              </p>
              <Button
                variant="ghost"
                size="icon"
                onClick={copyDeviceId}
                className="h-6 w-6 text-white/60 hover:text-white hover:bg-white/10"
                title="Copy Device ID"
              >
                {copied ? (
                  <Check className="h-3 w-3" />
                ) : (
                  <Copy className="h-3 w-3" />
                )}
              </Button>
            </div>
            {/* {piId ? (
              <p className="text-white/60 font-light text-sm">
                PI ID: {piId}
              </p>
            ) : isFindingPi ? (
              <div className="flex items-center gap-2 mt-2">
                <Loader2 className="h-4 w-4 text-white/60 animate-spin" />
                <p className="text-white/60 font-light text-sm">Finding device...</p>
              </div>
            ) : null} */}
          </div>

          {error && (
            <div className="mb-6 p-4 border border-red-500/20 bg-red-500/10 rounded-lg flex items-center gap-2">
              <AlertCircle className="h-5 w-5 text-red-400" />
              <p className="text-sm text-red-400 font-light">{error}</p>
            </div>
          )}

          {/* Latest Reading Card */}
          {latestReading && (
            <div className="mb-6 border border-white/10 rounded-lg p-6 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-xl font-light">Current Reading</h2>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={refreshSensorData}
                  disabled={isRefreshing}
                  className="h-8 w-8 text-orange-400 hover:text-orange-500 hover:bg-orange-500/10 disabled:opacity-50 border border-orange-400/90 border-2"
                  title="Refresh sensor data"
                > 
                  <RefreshCw className={`h-4 w-4 ${isRefreshing ? 'animate-spin' : ''}`} />
                </Button>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {latestReading.payload.sensors.temperature && (
                  <div className="border border-white/10 rounded-lg p-4 bg-white/5">
                    <div className="flex items-center gap-2 mb-2">
                      <Thermometer className="h-5 w-5 text-white/60" />
                      <span className="text-sm text-white/60 font-light">Temperature</span>
                    </div>
                    <div className="text-2xl font-light">
                      {latestReading.payload.sensors.temperature.value.toFixed(1)}
                      <span className="text-sm text-white/60 ml-1">
                        °{latestReading.payload.sensors.temperature.unit === "fahrenheit" || latestReading.payload.sensors.temperature.unit === "F" ? "F" : latestReading.payload.sensors.temperature.unit === "celsius" || latestReading.payload.sensors.temperature.unit === "C" ? "C" : latestReading.payload.sensors.temperature.unit.toUpperCase()}
                      </span>
                    </div>
                    <div className="text-xs text-white/40 font-light mt-1">
                      {formatDateShort(latestReading.ts)}
                    </div>
                  </div>
                )}
                {latestReading.payload.sensors.level && (
                  <div className="border border-white/10 rounded-lg p-4 bg-white/5">
                    <div className="flex items-center gap-2 mb-2">
                      <Droplets className="h-5 w-5 text-white/60" />
                      <span className="text-sm text-white/60 font-light">Sap Level</span>
                    </div>
                    <div className="text-2xl font-light">
                      {latestReading.payload.sensors.level.value.toFixed(1)}
                      <span className="text-sm text-white/60 ml-1">
                        {latestReading.payload.sensors.level.unit}
                      </span>
                    </div>
                    <div className="text-xs text-white/40 font-light mt-1">
                      {formatDateShort(latestReading.ts)}
                    </div>
                  </div>
                )}
                <div className="border border-white/10 rounded-lg p-4 bg-white/5">
                  <div className="flex items-center gap-2 mb-2">
                    <Battery className="h-5 w-5 text-white/60" />
                    <span className="text-sm text-white/60 font-light">Battery</span>
                  </div>
                  <div className="text-2xl font-light">
                    {latestReading.payload.battery_percentage.toFixed(1)}
                    <span className="text-sm text-white/60 ml-1">%</span>
                  </div>
                  <div className="text-xs text-white/40 font-light mt-1">
                    {formatDateShort(latestReading.ts)}
                  </div>
                </div>
              </div>
            </div>
          )}


          {/* Readings Chart */}
          {readings && Array.isArray(readings) && readings.length > 0 && (
            <div className="mb-6 border border-white/10 rounded-xl bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm overflow-hidden shadow-lg">
              <div className="px-6 py-5 border-b border-white/10 bg-gradient-to-r from-white/5 to-transparent">
                <div className="flex items-center justify-between mb-4">
                  <h2 className="text-2xl font-light text-white">Readings History</h2>
                  <div className="flex items-center gap-3">
                    <span className="text-sm text-white/70 font-light">Time Range:</span>
                    <div className="flex gap-2">
                      {(["1h", "1d", "1w", "1m", "1y"] as const).map((range) => (
                        <Button
                          key={range}
                          variant="ghost"
                          size="sm"
                          onClick={() => setTimeRange(range)}
                          className={`text-xs px-4 py-2 h-8 font-light transition-all ${
                            timeRange === range
                              ? "bg-orange-500 text-white hover:bg-orange-500/90 shadow-md shadow-orange-500/30"
                              : "text-white/70 hover:text-white hover:bg-orange-500/20 border border-orange-500/30"
                          }`}
                        >
                          {range === "1h" ? "1 Hour" : range === "1d" ? "1 Day" : range === "1w" ? "1 Week" : range === "1m" ? "1 Month" : "1 Year"}
                        </Button>
                      ))}
                    </div>
                  </div>
                </div>
              </div>
              <div className="p-6">
                <ReadingsChart readings={readings} timeRange={timeRange} />
              </div>
            </div>
          )}

          {/* Readings Table */}
          <div className="border border-white/10 rounded-lg bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm overflow-hidden">
            <div className="px-6 py-4 border-b border-white/10">
              <h2 className="text-xl font-light">Readings History</h2>
            </div>
            {piId ? (
              <ReadingsTable deviceId={deviceId} piId={piId} />
            ) : isFindingPi ? (
              <div className="p-12 flex items-center justify-center">
                <div className="flex flex-col items-center gap-3">
                  <Loader2 className="h-8 w-8 text-white/60 animate-spin" />
                  <p className="text-white/60 font-light">Loading readings...</p>
                </div>
              </div>
            ) : (
              <div className="p-12 text-center">
                <p className="text-white/60 font-light">Device information is being loaded...</p>
              </div>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}

