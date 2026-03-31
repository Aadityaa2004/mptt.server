"use client";

import { useState, useEffect, useCallback } from "react";
import { useSearchParams } from "next/navigation";
import Link from "next/link";
import { adminService } from "@/services/api/adminService";
import { ReadingsChart } from "@/components/sensors/ReadingsChart";
import {
  Loader2,
  AlertCircle,
  X,
  Trash2,
  RefreshCw,
  Thermometer,
  Droplets,
  Battery,
  ArrowLeft,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { Pi, Device, Reading } from "@/types/admin";

export default function AdminReadingsPage() {
  const searchParams = useSearchParams();
  const piFromUrl = searchParams.get("pi") ?? "";
  const deviceFromUrl = searchParams.get("device") ?? "";

  const [pis, setPis] = useState<Pi[]>([]);
  const [devices, setDevices] = useState<Device[]>([]);
  const [readings, setReadings] = useState<Reading[]>([]);
  const [selectedPiForReadings, setSelectedPiForReadings] = useState("");
  const [selectedDeviceForReadings, setSelectedDeviceForReadings] = useState<string | null>(null);
  const [showDeviceAnalytics, setShowDeviceAnalytics] = useState(false);
  const [deviceForReadings, setDeviceForReadings] = useState<Device | null>(null);
  const [latestReading, setLatestReading] = useState<Reading | null>(null);
  const [readingsForChart, setReadingsForChart] = useState<Reading[]>([]);
  const [isLoadingReadings, setIsLoadingReadings] = useState(false);
  const [timeRange, setTimeRange] = useState<"1h" | "1d" | "1w" | "1m" | "1y">("1d");
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [deleteRangeDevice, setDeleteRangeDevice] = useState<{ piId: string; deviceId: string | number } | null>(null);
  const [deleteRangeFrom, setDeleteRangeFrom] = useState("");
  const [deleteRangeTo, setDeleteRangeTo] = useState("");
  const [isDeletingRange, setIsDeletingRange] = useState(false);
  const [togglingCollectionKey, setTogglingCollectionKey] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (piFromUrl) setSelectedPiForReadings(piFromUrl);
    if (deviceFromUrl) {
      setSelectedDeviceForReadings(deviceFromUrl);
      setShowDeviceAnalytics(true);
    }
  }, [piFromUrl, deviceFromUrl]);

  const loadPis = useCallback(async () => {
    const data = await adminService.getAllPis(undefined, 1, 1000);
    setPis(data?.items || []);
  }, []);

  const loadDevicesForPi = useCallback(async (piId: string) => {
    if (!piId) { setDevices([]); return; }
    try {
      const data = await adminService.getDevices(piId, 1, 1000);
      setDevices(data?.items || []);
    } catch {
      setDevices([]);
    }
  }, []);

  const loadReadings = useCallback(async () => {
    if (!selectedPiForReadings) return;
    try {
      setLoading(true);
      setError(null);
      if (selectedDeviceForReadings) {
        const deviceId = /^\d+$/.test(selectedDeviceForReadings) ? parseInt(selectedDeviceForReadings, 10) : selectedDeviceForReadings;
        const data = await adminService.getDeviceReadings(selectedPiForReadings, deviceId);
        setReadings(data?.items || []);
      } else {
        const data = await adminService.getReadings({ pi_id: selectedPiForReadings });
        setReadings(data?.items || []);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load readings");
      setReadings([]);
    } finally {
      setLoading(false);
    }
  }, [selectedPiForReadings, selectedDeviceForReadings]);

  const loadDeviceAnalytics = useCallback(async () => {
    if (!selectedPiForReadings || !selectedDeviceForReadings) return;
    try {
      setIsLoadingReadings(true);
      setError(null);
      const deviceId = /^\d+$/.test(selectedDeviceForReadings) ? parseInt(selectedDeviceForReadings, 10) : selectedDeviceForReadings;
      const [deviceResp, latestResp, chartResp] = await Promise.all([
        adminService.getDevice(selectedPiForReadings, deviceId),
        adminService.getDeviceReadings(selectedPiForReadings, deviceId, { page: 1, page_size: 1 }),
        adminService.getDeviceReadings(selectedPiForReadings, deviceId, { page: 1, page_size: 100 }),
      ]);
      setDeviceForReadings(deviceResp);
      setLatestReading(latestResp?.items?.[0] ?? null);
      setReadingsForChart(chartResp?.items || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load device analytics");
      setDeviceForReadings(null);
      setLatestReading(null);
      setReadingsForChart([]);
    } finally {
      setIsLoadingReadings(false);
    }
  }, [selectedPiForReadings, selectedDeviceForReadings]);

  useEffect(() => {
    loadPis();
  }, [loadPis]);

  useEffect(() => {
    loadDevicesForPi(selectedPiForReadings);
  }, [selectedPiForReadings, loadDevicesForPi]);

  useEffect(() => {
    if (!showDeviceAnalytics) loadReadings();
  }, [selectedPiForReadings, selectedDeviceForReadings, showDeviceAnalytics, loadReadings]);

  useEffect(() => {
    if (showDeviceAnalytics && selectedPiForReadings && selectedDeviceForReadings) loadDeviceAnalytics();
  }, [showDeviceAnalytics, selectedPiForReadings, selectedDeviceForReadings, loadDeviceAnalytics]);

  const handleDeviceClick = (deviceId: string) => {
    setSelectedDeviceForReadings(deviceId);
    setShowDeviceAnalytics(true);
  };

  const handleBackToReadings = () => {
    setShowDeviceAnalytics(false);
    setSelectedDeviceForReadings(null);
    setLatestReading(null);
    setReadingsForChart([]);
    setDeviceForReadings(null);
  };

  const handleToggleCollection = async (piId: string, deviceId: string | number, currentEnabled: boolean) => {
    const key = `${piId}-${deviceId}`;
    try {
      setTogglingCollectionKey(key);
      setError(null);
      await adminService.updateDevice(piId, deviceId, { collection_enabled: !currentEnabled });
      await loadDeviceAnalytics();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update collection");
    } finally {
      setTogglingCollectionKey(null);
    }
  };

  const handleDeleteReadingsByRange = async () => {
    if (!deleteRangeDevice || !deleteRangeFrom || !deleteRangeTo) return;
    try {
      setIsDeletingRange(true);
      setError(null);
      const from = new Date(deleteRangeFrom).toISOString();
      const to = new Date(deleteRangeTo).toISOString();
      await adminService.deleteReadingsByDateRange(deleteRangeDevice.piId, deleteRangeDevice.deviceId, from, to);
      setDeleteRangeDevice(null);
      setDeleteRangeFrom("");
      setDeleteRangeTo("");
      await loadDeviceAnalytics();
      await loadReadings();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete readings");
    } finally {
      setIsDeletingRange(false);
    }
  };

  const handleRefresh = async () => {
    setError(null);
    if (showDeviceAnalytics) await loadDeviceAnalytics();
    else await loadReadings();
  };

  const refreshDeviceAnalytics = async () => {
    if (isRefreshing) return;
    setIsRefreshing(true);
    await loadDeviceAnalytics();
    setIsRefreshing(false);
  };

  return (
    <main className="pt-16 sm:pt-24 px-4 sm:px-6 lg:px-8 pb-12 sm:pb-16">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8 flex items-start justify-between">
          <div>
            <h1 className="text-4xl font-light tracking-tight mb-2">Reading Management</h1>
          </div>
          <button
            onClick={handleRefresh}
            disabled={loading || isLoadingReadings}
            className="p-2 rounded-lg border border-white/20 hover:bg-white/5 text-white/70 hover:text-white disabled:opacity-50 transition-colors"
            title="Refresh data"
          >
            <RefreshCw className={`h-4 w-4 ${loading || isLoadingReadings ? "animate-spin" : ""}`} />
          </button>
        </div>

        {error && (
          <div className="mb-6 p-4 border border-red-500/20 bg-red-500/10 rounded-lg flex items-center gap-2">
            <AlertCircle className="h-5 w-5 text-red-400" />
            <p className="text-sm text-red-400 font-light">{error}</p>
          </div>
        )}

        {!showDeviceAnalytics ? (
          <>
            <div className="mb-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <select
                  value={selectedPiForReadings}
                  onChange={(e) => {
                    setSelectedPiForReadings(e.target.value);
                    setSelectedDeviceForReadings(null);
                    setShowDeviceAnalytics(false);
                  }}
                  className="flex h-9 w-full rounded-md border border-white bg-black px-3 py-1 text-sm"
                >
                  <option value="">Select PI</option>
                  {pis.map((pi) => (
                    <option key={pi.pi_id} value={pi.pi_id}>{pi.pi_id}</option>
                  ))}
                </select>
                {selectedPiForReadings && (
                  <select
                    value={selectedDeviceForReadings || ""}
                    onChange={(e) => {
                      const id = e.target.value || null;
                      setSelectedDeviceForReadings(id);
                      if (id) handleDeviceClick(id);
                    }}
                    className="flex h-9 w-full rounded-md border border-white bg-black px-3 py-1 text-sm"
                  >
                    <option value="">All Devices</option>
                    {devices.filter((d) => d.pi_id === selectedPiForReadings).map((d) => (
                      <option key={d.device_id} value={String(d.device_id)}>Device {d.device_id}</option>
                    ))}
                  </select>
                )}
              </div>
            </div>

            {selectedPiForReadings && (
              <div className="border border-white/10 rounded-lg overflow-hidden">
                <table className="w-full">
                  <thead className="bg-black/50 border-b border-white/10">
                    <tr>
                      <th className="px-4 py-3 text-left text-sm font-light">Timestamp</th>
                      <th className="px-4 py-3 text-left text-sm font-light">Device ID</th>
                      <th className="px-4 py-3 text-left text-sm font-light">Temperature</th>
                      <th className="px-4 py-3 text-left text-sm font-light">Sap depth</th>
                      <th className="px-4 py-3 text-left text-sm font-light">Fill %</th>
                      <th className="px-4 py-3 text-left text-sm font-light">Battery</th>
                    </tr>
                  </thead>
                  <tbody>
                    {readings.map((r, i) => (
                      <tr key={i} className="border-b border-white/10 hover:bg-white/5">
                        <td className="px-4 py-3 text-sm font-light">{new Date(r.ts).toLocaleString()}</td>
                        <td className="px-4 py-3 text-sm font-light">
                          <button onClick={() => handleDeviceClick(String(r.device_id))} className="font-mono text-orange-400 hover:text-orange-300 hover:underline">
                            {r.device_id}
                          </button>
                        </td>
                        <td className="px-4 py-3 text-sm font-light">
                          {r.payload.sensors.temperature ? `${r.payload.sensors.temperature.value} ${r.payload.sensors.temperature.unit}` : "N/A"}
                        </td>
                        <td className="px-4 py-3 text-sm font-light">
                          {r.payload.sensors.level
                            ? r.sap_depth_cm != null
                              ? `${r.sap_depth_cm.toFixed(0)} cm sap`
                              : `${r.payload.sensors.level.value} cm to surface`
                            : "N/A"}
                        </td>
                        <td className="px-4 py-3 text-sm font-light">
                          {r.fill_percentage !== undefined ? <span className={r.fill_percentage >= 75 ? "text-orange-400 font-medium" : ""}>{r.fill_percentage.toFixed(1)}%</span> : "N/A"}
                        </td>
                        <td className="px-4 py-3 text-sm font-light">{r.payload.battery_percentage}%</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
                {readings.length === 0 && (
                  <div className="p-8 text-center text-white/60 font-light">No readings found.</div>
                )}
              </div>
            )}
          </>
        ) : (
          <>
            <div className="mb-6">
              <h2 className="text-2xl font-light mb-4">Device Analytics</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <select
                  value={selectedPiForReadings}
                  onChange={(e) => {
                    setSelectedPiForReadings(e.target.value);
                    setSelectedDeviceForReadings(null);
                    setShowDeviceAnalytics(false);
                  }}
                  className="flex h-9 w-full rounded-md border border-white bg-black px-3 py-1 text-sm"
                >
                  <option value="">Select PI</option>
                  {pis.map((pi) => (
                    <option key={pi.pi_id} value={pi.pi_id}>{pi.pi_id}</option>
                  ))}
                </select>
                {selectedPiForReadings && (
                  <select
                    value={selectedDeviceForReadings || ""}
                    onChange={(e) => {
                      const id = e.target.value || null;
                      setSelectedDeviceForReadings(id);
                      if (id) handleDeviceClick(id);
                      else setShowDeviceAnalytics(false);
                    }}
                    className="flex h-9 w-full rounded-md border border-white bg-black px-3 py-1 text-sm"
                  >
                    <option value="">All Devices</option>
                    {devices.filter((d) => d.pi_id === selectedPiForReadings).map((d) => (
                      <option key={d.device_id} value={String(d.device_id)}>Device {d.device_id}</option>
                    ))}
                  </select>
                )}
              </div>
            </div>

            <div className="mb-6">
              <Button variant="ghost" size="sm" onClick={handleBackToReadings} className="text-white/70 hover:text-white flex items-center gap-1">
                <ArrowLeft className="h-4 w-4" /> Back to readings list
              </Button>
            </div>

            {isLoadingReadings ? (
              <div className="flex justify-center py-12">
                <Loader2 className="h-6 w-6 text-white/60 animate-spin" />
              </div>
            ) : (
              <>
                {deviceForReadings && selectedPiForReadings && selectedDeviceForReadings && (
                  <div className="flex flex-wrap items-center gap-3 mb-6 p-4 border border-white/10 rounded-lg bg-black/30">
                    <span className="text-sm text-white/60 font-light">Collection:</span>
                    <button
                      onClick={() => handleToggleCollection(selectedPiForReadings, selectedDeviceForReadings, deviceForReadings.collection_enabled !== false)}
                      disabled={togglingCollectionKey === `${selectedPiForReadings}-${selectedDeviceForReadings}`}
                      className={`px-2 py-1 rounded text-xs font-light ${deviceForReadings.collection_enabled !== false ? "bg-green-500/20 text-green-400" : "bg-white/10 text-white/60"}`}
                    >
                      {togglingCollectionKey === `${selectedPiForReadings}-${selectedDeviceForReadings}` ? "..." : deviceForReadings.collection_enabled !== false ? "On" : "Off"}
                    </button>
                    <Button
                      size="sm"
                      variant="outline"
                      className="text-xs font-light h-7 px-3 text-orange-400 border-orange-400/50"
                      onClick={() => { setDeleteRangeDevice({ piId: selectedPiForReadings, deviceId: selectedDeviceForReadings }); setDeleteRangeFrom(""); setDeleteRangeTo(""); }}
                    >
                      Delete range
                    </Button>
                  </div>
                )}

                {latestReading && (
                  <div className="mb-6 border border-white/10 rounded-lg p-6 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm">
                    <div className="flex items-center justify-between mb-4">
                      <h2 className="text-xl font-light">Current Reading</h2>
                      <Button variant="ghost" size="icon" onClick={refreshDeviceAnalytics} disabled={isRefreshing} className="h-8 w-8 text-orange-400 hover:text-orange-500 hover:bg-orange-500/10 border border-orange-400/90" title="Refresh">
                        <RefreshCw className={`h-4 w-4 ${isRefreshing ? "animate-spin" : ""}`} />
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
                            <span className="text-sm text-white/60 ml-1">°</span>
                          </div>
                        </div>
                      )}
                      {latestReading.payload.sensors.level && (
                        <div className="border border-white/10 rounded-lg p-4 bg-white/5">
                          <div className="flex items-center gap-2 mb-2">
                            <Droplets className="h-5 w-5 text-white/60" />
                            <span className="text-sm text-white/60 font-light">Fill Level</span>
                          </div>
                          <div className="text-2xl font-light">
                            {latestReading.fill_percentage != null
                              ? `${latestReading.fill_percentage.toFixed(1)}% fill`
                              : latestReading.sap_depth_cm != null
                                ? `${latestReading.sap_depth_cm.toFixed(0)} cm sap`
                                : `${latestReading.payload.sensors.level.value.toFixed(1)} cm to surface`}
                          </div>
                        </div>
                      )}
                      <div className="border border-white/10 rounded-lg p-4 bg-white/5">
                        <div className="flex items-center gap-2 mb-2">
                          <Battery className="h-5 w-5 text-white/60" />
                          <span className="text-sm text-white/60 font-light">Battery</span>
                        </div>
                        <div className="text-2xl font-light">{latestReading.payload.battery_percentage.toFixed(1)}%</div>
                      </div>
                    </div>
                  </div>
                )}

                {readingsForChart.length > 0 && (
                  <div className="mb-6 border border-white/10 rounded-xl bg-gradient-to-br from-white/5 to-white/0 overflow-hidden">
                    <div className="px-6 py-5 border-b border-white/10">
                      <div className="flex items-center justify-between mb-4">
                        <h2 className="text-2xl font-light">Readings History</h2>
                        <div className="flex gap-2">
                          {(["1h", "1d", "1w", "1m", "1y"] as const).map((r) => (
                            <Button key={r} variant="ghost" size="sm" onClick={() => setTimeRange(r)}
                              className={`text-xs px-4 py-2 h-8 font-light ${timeRange === r ? "bg-orange-500 text-white" : "text-white/70 hover:bg-orange-500/20 border border-orange-500/30"}`}>
                              {r === "1h" ? "1 Hour" : r === "1d" ? "1 Day" : r === "1w" ? "1 Week" : r === "1m" ? "1 Month" : "1 Year"}
                            </Button>
                          ))}
                        </div>
                      </div>
                    </div>
                    <div className="p-6">
                      <ReadingsChart readings={readingsForChart} timeRange={timeRange} />
                    </div>
                  </div>
                )}

                {readingsForChart.length > 0 && (
                  <div className="border border-white/10 rounded-lg overflow-hidden">
                    <div className="px-6 py-4 border-b border-white/10 bg-black/50">
                      <h2 className="text-xl font-light">Readings Table</h2>
                    </div>
                    <table className="w-full">
                      <thead className="bg-black/50 border-b border-white/10">
                        <tr>
                          <th className="px-4 py-3 text-left text-sm font-light">Timestamp</th>
                          <th className="px-4 py-3 text-left text-sm font-light">Device ID</th>
                          <th className="px-4 py-3 text-left text-sm font-light">Temperature</th>
                          <th className="px-4 py-3 text-left text-sm font-light">Sap depth</th>
                          <th className="px-4 py-3 text-left text-sm font-light">Fill %</th>
                          <th className="px-4 py-3 text-left text-sm font-light">Battery</th>
                        </tr>
                      </thead>
                      <tbody>
                        {[...readingsForChart].sort((a, b) => new Date(b.ts).getTime() - new Date(a.ts).getTime()).map((r, i) => (
                          <tr key={i} className="border-b border-white/10 hover:bg-white/5">
                            <td className="px-4 py-3 text-sm font-light">{new Date(r.ts).toLocaleString()}</td>
                            <td className="px-4 py-3 text-sm font-light font-mono">{r.device_id}</td>
                            <td className="px-4 py-3 text-sm font-light">{r.payload.sensors.temperature ? `${r.payload.sensors.temperature.value} ${r.payload.sensors.temperature.unit}` : "N/A"}</td>
                            <td className="px-4 py-3 text-sm font-light">
                              {r.payload.sensors.level
                                ? r.sap_depth_cm != null
                                  ? `${r.sap_depth_cm.toFixed(0)} cm sap`
                                  : `${r.payload.sensors.level.value} cm to surface`
                                : "N/A"}
                            </td>
                            <td className="px-4 py-3 text-sm font-light">{r.fill_percentage !== undefined ? <span className={r.fill_percentage >= 75 ? "text-orange-400 font-medium" : ""}>{r.fill_percentage.toFixed(1)}%</span> : "N/A"}</td>
                            <td className="px-4 py-3 text-sm font-light">{r.payload.battery_percentage}%</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </>
            )}
          </>
        )}

        {deleteRangeDevice && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
            <div className="bg-black border border-white/10 rounded-lg p-6 max-w-md w-full mx-4">
              <h3 className="text-lg font-light mb-4">Delete readings by date range</h3>
              <p className="text-white/60 font-light mb-4 text-sm">Delete readings from start to end date (RFC3339-compatible).</p>
              <div className="space-y-3 mb-4">
                <Input type="datetime-local" value={deleteRangeFrom} onChange={(e) => setDeleteRangeFrom(e.target.value)} className="bg-black/50" />
                <Input type="datetime-local" value={deleteRangeTo} onChange={(e) => setDeleteRangeTo(e.target.value)} className="bg-black/50" />
              </div>
              <div className="flex gap-2">
                <Button variant="destructive" onClick={handleDeleteReadingsByRange} disabled={!deleteRangeFrom || !deleteRangeTo || isDeletingRange}>
                  {isDeletingRange ? "Deleting..." : "Delete"}
                </Button>
                <Button variant="outline" onClick={() => { setDeleteRangeDevice(null); setDeleteRangeFrom(""); setDeleteRangeTo(""); }}>Cancel</Button>
              </div>
            </div>
          </div>
        )}
      </div>
    </main>
  );
}
