"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { adminService } from "@/services/api/adminService";
import {
  Loader2,
  AlertCircle,
  X,
  Trash2,
  RefreshCw,
  CheckCircle2,
  XCircle,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { Pi, Device, Reading } from "@/types/admin";

const CONNECTED_WINDOW_MS = 15 * 60 * 1000;

function isValidMACAddress(mac: string): boolean {
  return /^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$/.test(mac);
}

function isValidDeviceId(id: string): boolean {
  if (!id?.trim()) return false;
  const t = id.trim();
  if (!isNaN(Number(t)) && Number(t) > 0) return true;
  if (isValidMACAddress(t)) return true;
  return t.length > 0;
}

export default function AdminDevicesPage() {
  const searchParams = useSearchParams();
  const piFilterFromUrl = searchParams.get("pi") ?? "";

  const [pis, setPis] = useState<Pi[]>([]);
  const [allDevices, setAllDevices] = useState<Device[]>([]);
  const [selectedPiFilter, setSelectedPiFilter] = useState("");
  const [latestReadingsByPi, setLatestReadingsByPi] = useState<Record<string, Reading[]>>({});
  const [showDeviceForm, setShowDeviceForm] = useState(false);
  const [showDeleteDeviceConfirm, setShowDeleteDeviceConfirm] = useState<{ piId: string; deviceId: number | string } | null>(null);
  const [deleteRangeDevice, setDeleteRangeDevice] = useState<{ piId: string; deviceId: string | number } | null>(null);
  const [deleteRangeFrom, setDeleteRangeFrom] = useState("");
  const [deleteRangeTo, setDeleteRangeTo] = useState("");
  const [isDeletingRange, setIsDeletingRange] = useState(false);
  const [togglingCollectionKey, setTogglingCollectionKey] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [deviceFormData, setDeviceFormData] = useState({
    device_id: "",
    height: "",
    top_diameter: "",
    bottom_diameter: "",
  });

  const selectedPiForForm = piFilterFromUrl ? { pi_id: piFilterFromUrl } : null;

  const isDeviceConnected = useCallback((piId: string, deviceId: number | string): boolean => {
    const latest = latestReadingsByPi[piId] || [];
    const r = latest.find((x) => String(x.device_id) === String(deviceId));
    if (!r?.ts) return false;
    return Date.now() - new Date(r.ts).getTime() <= CONNECTED_WINDOW_MS;
  }, [latestReadingsByPi]);

  const loadPis = useCallback(async () => {
    const data = await adminService.getAllPis(undefined, 1, 1000);
    setPis(data?.items || []);
    return data?.items || [];
  }, []);

  const loadAllDevices = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      let pisList: Pi[] = [];
      const pisData = await adminService.getAllPis(undefined, 1, 1000);
      pisList = pisData?.items || [];
      setPis(pisList);
      const results = await Promise.all(
        pisList.map(async (pi) => {
          try {
            const [devicesData, latestData] = await Promise.all([
              adminService.getDevices(pi.pi_id, 1, 1000),
              adminService.getLatestReadings(pi.pi_id).catch(() => ({ items: [] })),
            ]);
            return { devices: devicesData?.items || [], latest: latestData?.items || [], piId: pi.pi_id };
          } catch {
            return { devices: [], latest: [], piId: pi.pi_id };
          }
        })
      );
      setAllDevices(results.flatMap((r) => r.devices));
      const newLatest: Record<string, Reading[]> = {};
      results.forEach((r) => { newLatest[r.piId] = r.latest; });
      setLatestReadingsByPi((prev) => ({ ...prev, ...newLatest }));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load devices");
      setAllDevices([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (piFilterFromUrl) setSelectedPiFilter(piFilterFromUrl);
  }, [piFilterFromUrl]);

  useEffect(() => {
    loadAllDevices();
  }, [loadAllDevices]);

  const handleRefresh = async () => {
    setError(null);
    await loadPis();
    await loadAllDevices();
  };

  const handleToggleCollection = async (piId: string, deviceId: number | string, currentEnabled: boolean) => {
    const key = `${piId}-${deviceId}`;
    try {
      setTogglingCollectionKey(key);
      setError(null);
      await adminService.updateDevice(piId, deviceId, { collection_enabled: !currentEnabled });
      await loadAllDevices();
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
      await loadAllDevices();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete readings");
    } finally {
      setIsDeletingRange(false);
    }
  };

  const handleDeleteDevice = async (piId: string, deviceId: number | string) => {
    try {
      setLoading(true);
      setError(null);
      await adminService.deleteDevice(piId, deviceId, true);
      setShowDeleteDeviceConfirm(null);
      await loadAllDevices();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete device");
    } finally {
      setLoading(false);
    }
  };

  const handleCreateDevice = async () => {
    if (!selectedPiForForm) return;
    if (!deviceFormData.device_id || !isValidDeviceId(deviceFormData.device_id)) {
      setError("Device ID is required and must be a valid number or MAC address");
      return;
    }
    try {
      setLoading(true);
      setError(null);
      await adminService.createDevice(selectedPiForForm.pi_id, {
        device_id: deviceFormData.device_id.trim(),
        ...(deviceFormData.height ? { height: parseFloat(deviceFormData.height) } : {}),
        ...(deviceFormData.top_diameter ? { top_diameter: parseFloat(deviceFormData.top_diameter) } : {}),
        ...(deviceFormData.bottom_diameter ? { bottom_diameter: parseFloat(deviceFormData.bottom_diameter) } : {}),
      });
      setShowDeviceForm(false);
      setDeviceFormData({ device_id: "", height: "", top_diameter: "", bottom_diameter: "" });
      await loadAllDevices();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create device");
    } finally {
      setLoading(false);
    }
  };

  const filteredDevices =
    selectedPiFilter
      ? allDevices.filter((d) => d.pi_id === selectedPiFilter)
      : allDevices;

  return (
    <main className="pt-16 sm:pt-24 px-4 sm:px-6 lg:px-8 pb-12 sm:pb-16">
      <div className="max-w-7xl mx-auto">
        <div className="mb-6 sm:mb-8 flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
          <div>
            <h1 className="text-2xl sm:text-4xl font-light tracking-tight mb-2">
              {selectedPiForForm ? `Devices — PI: ${selectedPiForForm.pi_id}` : "Device Management"}
            </h1>
            <p className="text-white/60 font-light text-sm mt-1">
              {filteredDevices.length} device{filteredDevices.length !== 1 ? "s" : ""}
              {selectedPiFilter ? ` for PI: ${selectedPiFilter}` : " across all PIs"}
            </p>
          </div>
          <div className="flex gap-2 items-center">
            <button
              onClick={handleRefresh}
              disabled={loading}
              className="p-2 rounded-lg border border-white/20 hover:bg-white/5 text-white/70 hover:text-white disabled:opacity-50 transition-colors"
              title="Refresh data"
            >
              <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
            </button>
            {selectedPiForForm && (
              <Button onClick={() => setShowDeviceForm(true)}>Create Device</Button>
            )}
            {selectedPiFilter && (
              <Link href="/admin/devices">
                <Button variant="outline">Back to All Devices</Button>
              </Link>
            )}
          </div>
        </div>

        {!selectedPiForForm && (
          <div className="mb-6">
            <label className="text-sm text-white/60 font-light mr-2">Filter by PI:</label>
            <select
              value={selectedPiFilter}
              onChange={(e) => setSelectedPiFilter(e.target.value)}
              className="flex h-9 rounded-md border border-white bg-black px-3 py-1 text-sm min-w-[200px]"
            >
              <option value="">All PIs</option>
              {pis.map((pi) => (
                <option key={pi.pi_id} value={pi.pi_id}>{pi.pi_id}</option>
              ))}
            </select>
          </div>
        )}

        {error && (
          <div className="mb-6 p-4 border border-red-500/20 bg-red-500/10 rounded-lg flex items-center gap-2">
            <AlertCircle className="h-5 w-5 text-red-400" />
            <p className="text-sm text-red-400 font-light">{error}</p>
          </div>
        )}

        {showDeviceForm && selectedPiForForm && (
          <div className="border border-white/10 rounded-lg p-6 bg-black/50 mb-6">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-light">Create New Device</h3>
              <button onClick={() => setShowDeviceForm(false)} className="text-white/60 hover:text-white">
                <X className="h-5 w-5" />
              </button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="text-sm text-white/60 font-light mb-1 block">Device ID</label>
                <Input
                  placeholder="Number or MAC (e.g., AA:BB:CC:DD:EE:FF)"
                  value={deviceFormData.device_id}
                  onChange={(e) => setDeviceFormData({ ...deviceFormData, device_id: e.target.value })}
                  className="bg-black/50 border-white/10"
                />
              </div>
              <div>
                <label className="text-sm text-white/60 font-light mb-1 block">Bucket Dimensions (optional, cm)</label>
                <div className="grid grid-cols-3 gap-3">
                  <Input type="number" placeholder="Height" value={deviceFormData.height} onChange={(e) => setDeviceFormData({ ...deviceFormData, height: e.target.value })} className="bg-black/50 border-white/10" />
                  <Input type="number" placeholder="Top Diameter" value={deviceFormData.top_diameter} onChange={(e) => setDeviceFormData({ ...deviceFormData, top_diameter: e.target.value })} className="bg-black/50 border-white/10" />
                  <Input type="number" placeholder="Bottom Diameter" value={deviceFormData.bottom_diameter} onChange={(e) => setDeviceFormData({ ...deviceFormData, bottom_diameter: e.target.value })} className="bg-black/50 border-white/10" />
                </div>
              </div>
            </div>
            <div className="flex gap-2 mt-4">
              <Button onClick={handleCreateDevice} disabled={!deviceFormData.device_id || !isValidDeviceId(deviceFormData.device_id)}>
                Create
              </Button>
              <Button variant="outline" onClick={() => setShowDeviceForm(false)}>Cancel</Button>
            </div>
          </div>
        )}

        {loading && !allDevices.length ? (
          <div className="flex justify-center py-12">
            <Loader2 className="h-6 w-6 text-white/60 animate-spin" />
          </div>
        ) : filteredDevices.length === 0 ? (
          <div className="border border-white/10 rounded-lg p-12 bg-black/50 text-center">
            <p className="text-white/60 font-light mb-4">No devices found</p>
            <p className="text-white/40 font-light text-sm">
              {selectedPiFilter ? `No devices for PI: ${selectedPiFilter}` : "Devices will appear once registered to a PI."}
            </p>
          </div>
        ) : (
          <div className="border border-white/10 rounded-lg overflow-hidden">
            <table className="w-full">
              <thead className="bg-black/50 border-b border-white/10">
                <tr>
                  <th className="px-4 py-3 text-left text-sm font-light">PI ID</th>
                  <th className="px-4 py-3 text-left text-sm font-light">Device ID</th>
                  <th className="px-4 py-3 text-left text-sm font-light">Status</th>
                  <th className="px-4 py-3 text-left text-sm font-light">Collection</th>
                  <th className="px-4 py-3 text-left text-sm font-light">View</th>
                  <th className="px-4 py-3 text-left text-sm font-light">Delete data</th>
                  <th className="px-4 py-3 text-left text-sm font-light">Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredDevices.map((device) => {
                  const key = `${device.pi_id}-${device.device_id}`;
                  const toggling = togglingCollectionKey === key;
                  const enabled = device.collection_enabled !== false;
                  const connected = isDeviceConnected(device.pi_id, device.device_id);
                  return (
                    <tr key={key} className="border-b border-white/10 hover:bg-white/5">
                      <td className="px-4 py-3 text-sm font-light font-mono">{device.pi_id}</td>
                      <td className="px-4 py-3 text-sm font-light font-mono">{device.device_id}</td>
                      <td className="px-4 py-3 text-sm font-light">
                        {connected ? (
                          <span className="inline-flex items-center gap-1 text-green-400"><CheckCircle2 className="h-4 w-4" /> Connected</span>
                        ) : (
                          <span className="inline-flex items-center gap-1 text-white/50"><XCircle className="h-4 w-4" /> Disconnected</span>
                        )}
                      </td>
                      <td className="px-4 py-3 text-sm font-light">
                        <button
                          onClick={() => handleToggleCollection(device.pi_id, device.device_id, enabled)}
                          disabled={toggling}
                          className={`px-2 py-1 rounded text-xs font-light ${enabled ? "bg-green-500/20 text-green-400" : "bg-white/10 text-white/60"}`}
                        >
                          {toggling ? "..." : enabled ? "On" : "Off"}
                        </button>
                      </td>
                      <td className="px-4 py-3 text-sm font-light">
                        <Link href={`/admin/readings?pi=${encodeURIComponent(device.pi_id)}&device=${encodeURIComponent(String(device.device_id))}`}>
                          <Button size="sm" variant="outline" className="text-xs font-light h-7 px-3">View Readings</Button>
                        </Link>
                      </td>
                      <td className="px-4 py-3 text-sm font-light">
                        <Button
                          size="sm"
                          variant="outline"
                          className="text-xs font-light h-7 px-3 text-orange-400 border-orange-400/50"
                          onClick={() => { setDeleteRangeDevice({ piId: device.pi_id, deviceId: device.device_id }); setDeleteRangeFrom(""); setDeleteRangeTo(""); }}
                        >
                          Delete range
                        </Button>
                      </td>
                      <td className="px-4 py-3 text-sm font-light">
                        <button onClick={() => setShowDeleteDeviceConfirm({ piId: device.pi_id, deviceId: device.device_id })} className="text-red-400 hover:text-red-300" title="Delete device">
                          <Trash2 className="h-4 w-4" />
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}

        {showDeleteDeviceConfirm && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
            <div className="bg-black border border-white/10 rounded-lg p-6 max-w-md w-full mx-4">
              <h3 className="text-lg font-light mb-4">Confirm Delete</h3>
              <p className="text-white/60 font-light mb-6">Delete this device and all associated readings (cascade)?</p>
              <div className="flex gap-2">
                <Button variant="destructive" onClick={() => handleDeleteDevice(showDeleteDeviceConfirm.piId, showDeleteDeviceConfirm.deviceId)}>Delete</Button>
                <Button variant="outline" onClick={() => setShowDeleteDeviceConfirm(null)}>Cancel</Button>
              </div>
            </div>
          </div>
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
