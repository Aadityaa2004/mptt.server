"use client";

import { useMemo } from "react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { Droplets, Battery, Thermometer } from "lucide-react";
import type { Reading } from "@/types/admin";

interface ReadingsChartProps {
  readings: Reading[];
  timeRange: "1h" | "1d" | "1w" | "1m" | "1y";
}

interface CustomTooltipPayloadEntry {
  value: number | null;
}

interface ReadingsChartTooltipProps {
  active?: boolean;
  payload?: CustomTooltipPayloadEntry[];
  label?: string;
  unit: string;
}

// Format time label based on the time range
function formatTimeLabel(timestamp: string, range: "1h" | "1d" | "1w" | "1m" | "1y") {
  const date = new Date(timestamp);
  switch (range) {
    case "1h":
      return date.toLocaleString("en-US", { hour: "2-digit", minute: "2-digit" });
    case "1d":
      return date.toLocaleString("en-US", { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" });
    case "1w":
      return date.toLocaleString("en-US", { month: "short", day: "numeric", hour: "2-digit" });
    case "1m":
      return date.toLocaleString("en-US", { month: "short", day: "numeric" });
    case "1y":
      return date.toLocaleString("en-US", { month: "short", year: "numeric" });
    default:
      return date.toLocaleString("en-US", { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" });
  }
}

function getCutoffTime(latestTimestamp: number, timeRange: "1h" | "1d" | "1w" | "1m" | "1y") {
  switch (timeRange) {
    case "1h":
      return latestTimestamp - 60 * 60 * 1000;
    case "1d":
      return latestTimestamp - 24 * 60 * 60 * 1000;
    case "1w":
      return latestTimestamp - 7 * 24 * 60 * 60 * 1000;
    case "1m":
      return latestTimestamp - 30 * 24 * 60 * 60 * 1000;
    case "1y":
      return latestTimestamp - 365 * 24 * 60 * 60 * 1000;
    default:
      return latestTimestamp;
  }
}

function ReadingsChartTooltip({ active, payload, label, unit }: ReadingsChartTooltipProps) {
  if (!active || !payload || !payload.length) {
    return null;
  }

  const value = payload[0].value;
  if (value === null || value === undefined) {
    return null;
  }

  return (
    <div className="bg-black/95 border border-orange-500/30 rounded-lg p-3 shadow-xl shadow-orange-500/20">
      <p className="text-white/70 text-xs mb-2 font-light">{label}</p>
      <p className="text-base font-light text-orange-400">
        {`${value.toFixed(1)}${unit}`}
      </p>
    </div>
  );
}

export function ReadingsChart({ readings, timeRange }: ReadingsChartProps) {
  const chartData = useMemo(() => {
    if (readings.length === 0) {
      return [];
    }

    // Find the most recent timestamp in the dataset
    const timestamps = readings.map((reading) => new Date(reading.ts).getTime());
    const latestTimestamp = Math.max(...timestamps);
    
    // Calculate cutoff time based on the latest timestamp
    const cutoffTimestamp = getCutoffTime(latestTimestamp, timeRange);
    
    const filtered = readings
      .filter((reading) => {
        const readingTime = new Date(reading.ts).getTime();
        return readingTime >= cutoffTimestamp;
      })
      .sort((a, b) => {
        // Sort by timestamp ascending (oldest first)
        return new Date(a.ts).getTime() - new Date(b.ts).getTime();
      })
      .map((reading) => {
        const timeLabel = formatTimeLabel(reading.ts, timeRange);
        // Prefer sap_depth_cm (fill from bottom) for user-friendly chart
        const sapValue =
          reading.sap_depth_cm != null
            ? reading.sap_depth_cm
            : reading.payload.sensors.level?.value ?? null;

        return {
          timestamp: reading.ts,
          timeLabel,
          temperature: reading.payload.sensors.temperature?.value ?? null,
          sapLevel: sapValue,
          battery: reading.payload.battery_percentage ?? null,
        };
      });
    
    return filtered;
  }, [readings, timeRange]);

  if (readings.length === 0) {
    return null;
  }

  const hasTemperature = chartData.some((d) => d.temperature !== null);
  const hasSapLevel = chartData.some((d) => d.sapLevel !== null);
  const hasBattery = chartData.some((d) => d.battery !== null);

  // Calculate X-axis interval to show approximately 8-12 labels
  const getXAxisInterval = () => {
    if (chartData.length <= 12) return 0; // Show all labels if 12 or fewer
    // Show every Nth label to get ~10 labels total
    return Math.floor(chartData.length / 10);
  };

  const xAxisInterval = getXAxisInterval();

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
      {/* Temperature Chart */}
      {hasTemperature && (
        <div className="rounded-2xl p-6 bg-white/[0.04] border border-white/5">
          <div className="flex items-center gap-3 mb-4">
            <div className="p-2 rounded-xl bg-orange-500/20">
              <Thermometer className="h-5 w-5 text-orange-400/90" />
            </div>
            <div>
              <h3 className="text-base font-light text-white/90">Temperature</h3>
              <p className="text-xs text-white/50 font-light">Over time</p>
            </div>
          </div>
          <div className="w-full h-[300px]">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart
                key={`temperature-${timeRange}`}
                data={chartData}
                margin={{ top: 10, right: 20, left: 10, bottom: timeRange === "1h" ? 30 : 60 }}
              >
                <CartesianGrid strokeDasharray="3 3" stroke="#ffffff10" />
                <XAxis
                  dataKey="timeLabel"
                  stroke="#ffffff50"
                  style={{ fontSize: "11px" }}
                  angle={timeRange === "1h" ? 0 : -45}
                  textAnchor={timeRange === "1h" ? "middle" : "end"}
                  height={timeRange === "1h" ? 30 : 60}
                  interval={xAxisInterval}
                />
                <YAxis
                  stroke="#ffffff50"
                  style={{ fontSize: "12px" }}
                  label={{ value: "°C", angle: -90, position: "insideLeft", style: { fill: "#ffffff70", fontSize: "12px" } }}
                />
                <Tooltip 
                  content={<ReadingsChartTooltip unit="°C" />}
                  cursor={{ stroke: "#f97316", strokeWidth: 1, strokeDasharray: "5 5" }}
                />
                <Line
                  type="monotone"
                  dataKey="temperature"
                  stroke="#f97316"
                  strokeWidth={3}
                  dot={false}
                  connectNulls={false}
                  activeDot={{ r: 5, fill: "#ea580c", stroke: "#fff", strokeWidth: 2 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>
      )}

      {/* Sap Level Chart */}
      {hasSapLevel && (
        <div className="rounded-2xl p-6 bg-white/[0.04] border border-white/5">
          <div className="flex items-center gap-3 mb-4">
            <div className="p-2 rounded-xl bg-orange-500/20">
              <Droplets className="h-5 w-5 text-orange-400/90" />
            </div>
            <div>
              <h3 className="text-base font-light text-white/90">Sap Depth</h3>
              <p className="text-xs text-white/50 font-light">Fill level over time (cm from bottom)</p>
            </div>
          </div>
          <div className="w-full h-[300px]">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart
                key={`sapLevel-${timeRange}`}
                data={chartData}
                margin={{ top: 10, right: 20, left: 10, bottom: timeRange === "1h" ? 30 : 60 }}
              >
                <CartesianGrid strokeDasharray="3 3" stroke="#ffffff10" />
                <XAxis
                  dataKey="timeLabel"
                  stroke="#ffffff50"
                  style={{ fontSize: "11px" }}
                  angle={timeRange === "1h" ? 0 : -45}
                  textAnchor={timeRange === "1h" ? "middle" : "end"}
                  height={timeRange === "1h" ? 30 : 60}
                  interval={xAxisInterval}
                />
                <YAxis
                  stroke="#ffffff50"
                  style={{ fontSize: "12px" }}
                  label={{ value: "cm", angle: -90, position: "insideLeft", style: { fill: "#ffffff70", fontSize: "12px" } }}
                />
                <Tooltip 
                  content={<ReadingsChartTooltip unit=" cm" />}
                  cursor={{ stroke: "#f97316", strokeWidth: 1, strokeDasharray: "5 5" }}
                />
                <Line
                  type="monotone"
                  dataKey="sapLevel"
                  stroke="#f97316"
                  strokeWidth={3}
                  dot={false}
                  connectNulls={false}
                  activeDot={{ r: 5, fill: "#ea580c", stroke: "#fff", strokeWidth: 2 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>
      )}

      {/* Battery Chart */}
      {hasBattery && (
        <div className="rounded-2xl p-6 bg-white/[0.04] border border-white/5">
          <div className="flex items-center gap-3 mb-4">
            <div className="p-2 rounded-xl bg-orange-500/20">
              <Battery className="h-5 w-5 text-orange-400/90" />
            </div>
            <div>
              <h3 className="text-base font-light text-white/90">Battery</h3>
              <p className="text-xs text-white/50 font-light">Over time</p>
            </div>
          </div>
          <div className="w-full h-[300px]">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart
                key={`battery-${timeRange}`}
                data={chartData}
                margin={{ top: 10, right: 20, left: 10, bottom: timeRange === "1h" ? 30 : 60 }}
              >
                <CartesianGrid strokeDasharray="3 3" stroke="#ffffff10" />
                <XAxis
                  dataKey="timeLabel"
                  stroke="#ffffff50"
                  style={{ fontSize: "11px" }}
                  angle={timeRange === "1h" ? 0 : -45}
                  textAnchor={timeRange === "1h" ? "middle" : "end"}
                  height={timeRange === "1h" ? 30 : 60}
                  interval={xAxisInterval}
                />
                <YAxis
                  stroke="#ffffff50"
                  style={{ fontSize: "12px" }}
                  label={{ value: "%", angle: -90, position: "insideLeft", style: { fill: "#ffffff70", fontSize: "12px" } }}
                />
                <Tooltip 
                  content={<ReadingsChartTooltip unit="%" />}
                  cursor={{ stroke: "#ea580c", strokeWidth: 1, strokeDasharray: "5 5" }}
                />
                <Line
                  type="monotone"
                  dataKey="battery"
                  stroke="#ea580c"
                  strokeWidth={3}
                  dot={false}
                  connectNulls={false}
                  activeDot={{ r: 5, fill: "#c2410c", stroke: "#fff", strokeWidth: 2 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>
      )}
    </div>
  );
}

