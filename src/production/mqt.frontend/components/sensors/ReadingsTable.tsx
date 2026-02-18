"use client";

import { useState, useEffect } from "react";
import type { Reading } from "@/types/admin";
import { Loader2, Thermometer, Droplets, Battery } from "lucide-react";
import { Button } from "@/components/ui/button";

interface ReadingsTableProps {
  deviceId: string;
  piId: string;
  initialPage?: number;
  limit?: number;
}

export function ReadingsTable({ deviceId, piId, initialPage = 1, limit = 20 }: ReadingsTableProps) {
  const [readings, setReadings] = useState<Reading[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(initialPage);
  const [hasMore, setHasMore] = useState(false);
  const [nextPageToken, setNextPageToken] = useState<string | null>(null);

  const loadReadings = async (pageNum: number, append = false) => {
    try {
      setIsLoading(true);
      setError(null);

      const queryParams = new URLSearchParams({
        device_id: deviceId,
        pi_id: piId,
        page: pageNum.toString(),
        limit: limit.toString(),
      });

      // Get access token from sessionStorage for the API call
      const token = typeof window !== "undefined" ? sessionStorage.getItem("access_token") : null;
      
      const headers: HeadersInit = {
        "Content-Type": "application/json",
      };
      
      if (token) {
        headers["Authorization"] = `Bearer ${token}`;
      }

      // Use query parameter instead of path parameter to avoid encoding issues
      const url = `/api/user/sensors/readings?${queryParams.toString()}`;
      
      const response = await fetch(url, {
        headers,
        credentials: "include",
      });
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ error: "Failed to fetch readings" }));
        throw new Error(errorData.error || "Failed to fetch readings");
      }

      const data = await response.json();
      
      // Handle empty response
      if (!data || !data.items || !Array.isArray(data.items) || data.items.length === 0) {
        if (!append) {
          setReadings([]);
        }
        setHasMore(false);
        setNextPageToken(null);
        return;
      }

      if (append) {
        setReadings((prev) => [...prev, ...data.items]);
      } else {
        setReadings(data.items);
      }

      setNextPageToken(data.next_page_token || null);
      setHasMore(data.next_page_token !== null && data.next_page_token !== undefined);
    } catch (err) {
      console.error("Error loading readings:", err);
      const errorMessage = err instanceof Error ? err.message : "Failed to load readings";
      setError(errorMessage);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (piId && deviceId) {
      loadReadings(page, page > 1);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [piId, deviceId, page]);

  const handleLoadMore = () => {
    setPage((p) => p + 1);
  };

  const handleRefresh = () => {
    setPage(1);
    setReadings([]);
    loadReadings(1, false);
  };

  const formatDate = (timestamp: string) => {
    return new Date(timestamp).toLocaleString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    });
  };

  if (isLoading && readings.length === 0) {
    return (
      <div className="p-12 flex items-center justify-center">
        <div className="flex flex-col items-center gap-3">
          <Loader2 className="h-8 w-8 text-white/60 animate-spin" />
          <p className="text-white/60 font-light">Loading readings...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-12 text-center">
        <p className="text-red-400 font-light mb-4">{error}</p>
        <Button
          variant="outline"
          onClick={handleRefresh}
          className="text-white border-white/20 hover:bg-white/10"
        >
          Retry
        </Button>
      </div>
    );
  }

  if (readings.length === 0) {
    return (
      <div className="p-12 text-center">
        <p className="text-white/60 font-light mb-2">No data shown yet</p>
        <p className="text-white/40 font-light text-sm">
          Readings will appear here once the sensor starts sending data
        </p>
        <Button
          variant="outline"
          onClick={handleRefresh}
          className="mt-4 text-white border-white/20 hover:bg-white/10"
        >
          Refresh
        </Button>
      </div>
    );
  }

  return (
    <>
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="border-b border-white/10">
              <th className="px-6 py-3 text-left text-sm font-light text-white/60">Timestamp</th>
              <th className="px-6 py-3 text-left text-sm font-light text-white/60">Temperature</th>
              <th className="px-6 py-3 text-left text-sm font-light text-white/60">Sap Level</th>
              <th className="px-6 py-3 text-left text-sm font-light text-white/60">Battery</th>
            </tr>
          </thead>
          <tbody>
            {readings.map((reading, idx) => (
              <tr
                key={`${reading.ts}-${idx}`}
                className="border-b border-white/10 hover:bg-white/5 transition-colors"
              >
                <td className="px-6 py-4 text-sm font-light">
                  {formatDate(reading.ts)}
                </td>
                <td className="px-6 py-4 text-sm font-light">
                  {reading.payload.sensors.temperature ? (
                    <div className="flex items-center gap-2">
                      <Thermometer className="h-4 w-4 text-white/40" />
                      <span>
                        {reading.payload.sensors.temperature.value.toFixed(1)}°
                        {reading.payload.sensors.temperature.unit === "fahrenheit" || reading.payload.sensors.temperature.unit === "F" 
                          ? "F" 
                          : reading.payload.sensors.temperature.unit === "celsius" || reading.payload.sensors.temperature.unit === "C" 
                          ? "C" 
                          : reading.payload.sensors.temperature.unit.toUpperCase()}
                      </span>
                    </div>
                  ) : (
                    <span className="text-white/40">N/A</span>
                  )}
                </td>
                <td className="px-6 py-4 text-sm font-light">
                  {reading.payload.sensors.level ? (
                    <div className="flex items-center gap-2">
                      <Droplets className="h-4 w-4 text-white/40" />
                      <span>
                        {reading.payload.sensors.level.value.toFixed(1)} {reading.payload.sensors.level.unit}
                      </span>
                    </div>
                  ) : (
                    <span className="text-white/40">N/A</span>
                  )}
                </td>
                <td className="px-6 py-4 text-sm font-light">
                  <div className="flex items-center gap-2">
                    <Battery className="h-4 w-4 text-white/40" />
                    <span>{reading.payload.battery_percentage}%</span>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {hasMore && (
        <div className="px-6 py-4 border-t border-white/10 text-center">
          <Button
            variant="outline"
            onClick={handleLoadMore}
            disabled={isLoading}
            className="text-white border-white/20 hover:bg-white/10"
          >
            {isLoading ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Loading...
              </>
            ) : (
              "Load More"
            )}
          </Button>
        </div>
      )}
    </>
  );
}

