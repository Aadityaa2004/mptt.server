"use client";

import { useState, useEffect, useCallback } from "react";
import dynamic from "next/dynamic";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { adminService } from "@/services/api/adminService";
import { ReadingsChart } from "@/components/sensors/ReadingsChart";
import { Loader2, AlertCircle, ArrowLeft, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { User, Pi, Device, Reading } from "@/types/admin";

const UserDetailMap = dynamic(
  () => import("./UserDetailMap").then((m) => m.UserDetailMap),
  { ssr: false }
);

export default function AdminUserDetailPage() {
  const params = useParams();
  const router = useRouter();
  const userId = params.userId as string;

  const [user, setUser] = useState<User | null>(null);
  const [pis, setPis] = useState<Pi[]>([]);
  const [devicesByPi, setDevicesByPi] = useState<Record<string, Device[]>>({});
  const [readingsByDevice, setReadingsByDevice] = useState<Record<string, Reading[]>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadUserDetail = useCallback(async () => {
    if (!userId) return;
    try {
      setLoading(true);
      setError(null);
      const [userData, pisData] = await Promise.all([
        adminService.getUserById(userId),
        adminService.getAllPis(userId, 1, 100),
      ]);
      setUser(userData);
      const items = pisData?.items || [];
      setPis(items);
      const deviceMap: Record<string, Device[]> = {};
      await Promise.all(
        items.map(async (pi) => {
          try {
            const devData = await adminService.getDevices(pi.pi_id, 1, 100);
            deviceMap[pi.pi_id] = devData?.items || [];
          } catch {
            deviceMap[pi.pi_id] = [];
          }
        })
      );
      setDevicesByPi(deviceMap);

      // Fetch readings for each device
      const readingsMap: Record<string, Reading[]> = {};
      for (const pi of items) {
        const devs = deviceMap[pi.pi_id] || [];
        for (const dev of devs) {
          const key = `${pi.pi_id}-${dev.device_id}`;
          try {
            const resp = await adminService.getDeviceReadings(pi.pi_id, dev.device_id, { page: 1, page_size: 100 });
            readingsMap[key] = resp?.items || [];
          } catch {
            readingsMap[key] = [];
          }
        }
      }
      setReadingsByDevice(readingsMap);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load user");
      setUser(null);
      setPis([]);
      setDevicesByPi({});
      setReadingsByDevice({});
    } finally {
      setLoading(false);
    }
  }, [userId]);

  useEffect(() => {
    loadUserDetail();
  }, [loadUserDetail]);

  const handleRefresh = async () => {
    setError(null);
    await loadUserDetail();
  };

  if (!userId) {
    router.replace("/admin/users");
    return null;
  }

  if (loading && !user) {
    return (
      <main className="pt-24 px-4 sm:px-6 lg:px-8 pb-16">
        <div className="max-w-7xl mx-auto flex justify-center py-12">
          <Loader2 className="h-6 w-6 text-white/60 animate-spin" />
        </div>
      </main>
    );
  }

  if (error && !user) {
    return (
      <main className="pt-24 px-4 sm:px-6 lg:px-8 pb-16">
        <div className="max-w-7xl mx-auto">
          <div className="mb-6 p-4 border border-red-500/20 bg-red-500/10 rounded-lg flex items-center gap-2">
            <AlertCircle className="h-5 w-5 text-red-400" />
            <p className="text-sm text-red-400 font-light">{error}</p>
          </div>
          <Link href="/admin/users">
            <Button variant="outline">Back to Users</Button>
          </Link>
        </div>
      </main>
    );
  }

  return (
    <main className="pt-24 px-4 sm:px-6 lg:px-8 pb-16">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8 flex items-start justify-between">
          <div>
            <Link href="/admin/users" className="text-white/60 hover:text-white text-sm font-light flex items-center gap-1 mb-2">
              <ArrowLeft className="h-4 w-4" /> Back to Users
            </Link>
            <h1 className="text-4xl font-light tracking-tight mb-2">
              {user?.username ?? "User"}
            </h1>
            <p className="text-white/60 font-light text-sm">{user?.email}</p>
          </div>
          <button
            onClick={handleRefresh}
            disabled={loading}
            className="p-2 rounded-lg border border-white/20 hover:bg-white/5 text-white/70 hover:text-white disabled:opacity-50 transition-colors"
            title="Refresh data"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
          </button>
        </div>

        {error && (
          <div className="mb-6 p-4 border border-red-500/20 bg-red-500/10 rounded-lg flex items-center gap-2">
            <AlertCircle className="h-5 w-5 text-red-400" />
            <p className="text-sm text-red-400 font-light">{error}</p>
          </div>
        )}

        <div className="space-y-6">
          <div className="border border-white/10 rounded-lg p-6 bg-black/30">
            <h2 className="text-lg font-light mb-4">User Info</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <span className="text-white/60 text-sm font-light">Role</span>
                <p className="font-light">{user?.role ?? "—"}</p>
              </div>
              <div>
                <span className="text-white/60 text-sm font-light">Status</span>
                <p className="font-light">{user?.active ? "Active" : "Inactive"}</p>
              </div>
            </div>
          </div>

          <div className="border border-white/10 rounded-lg p-6 bg-black/30">
            <h2 className="text-lg font-light mb-4">PIs and Devices</h2>
            {pis.length === 0 ? (
              <p className="text-white/60 font-light">No PIs assigned to this user.</p>
            ) : (
              <div className="space-y-4">
                {pis.map((pi) => (
                  <div key={pi.pi_id} className="border border-white/10 rounded-lg p-4">
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-mono text-orange-400">{pi.pi_id}</span>
                      <Link href={`/admin/devices?pi=${encodeURIComponent(pi.pi_id)}`}>
                        <Button size="sm" variant="outline" className="text-xs">View Devices</Button>
                      </Link>
                    </div>
                    <div className="text-sm text-white/60 font-light">
                      {(devicesByPi[pi.pi_id] || []).length} device{(devicesByPi[pi.pi_id] || []).length !== 1 ? "s" : ""}
                    </div>
                    {(devicesByPi[pi.pi_id] || []).length > 0 && (
                      <div className="mt-2 flex flex-wrap gap-2">
                        {(devicesByPi[pi.pi_id] || []).map((d) => (
                          <Link key={d.device_id} href={`/admin/readings?pi=${encodeURIComponent(pi.pi_id)}&device=${encodeURIComponent(String(d.device_id))}`}>
                            <span className="text-xs px-2 py-1 rounded bg-white/10 hover:bg-white/20 transition-colors">
                              Device {d.device_id}
                            </span>
                          </Link>
                        ))}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Per-device readings charts */}
          {(() => {
            const charts: React.ReactNode[] = [];
            pis.forEach((pi) => {
              (devicesByPi[pi.pi_id] || []).forEach((d) => {
                const key = `${pi.pi_id}-${d.device_id}`;
                const readings = readingsByDevice[key] || [];
                if (readings.length > 0) {
                  charts.push(
                    <div key={key} className="border border-white/10 rounded-lg p-4">
                      <div className="flex items-center justify-between mb-3">
                        <span className="font-mono text-orange-400">PI: {pi.pi_id} / Device: {d.device_id}</span>
                        <Link href={`/admin/readings?pi=${encodeURIComponent(pi.pi_id)}&device=${encodeURIComponent(String(d.device_id))}`}>
                          <Button size="sm" variant="outline" className="text-xs">View full readings</Button>
                        </Link>
                      </div>
                      <ReadingsChart readings={readings} timeRange="1d" />
                    </div>
                  );
                }
              });
            });
            return charts.length > 0 ? (
              <div className="border border-white/10 rounded-lg p-6 bg-black/30">
                <h2 className="text-lg font-light mb-4">Device Readings</h2>
                <div className="space-y-6">{charts}</div>
              </div>
            ) : null;
          })()}

          {/* Map when user has device locations */}
          {user?.locations && user.locations.length > 0 && (
            <div className="border border-white/10 rounded-lg p-6 bg-black/30">
              <h2 className="text-lg font-light mb-4">Device Map</h2>
              <div className="h-[400px] rounded-lg overflow-hidden">
                <UserDetailMap locations={user.locations} />
              </div>
            </div>
          )}
        </div>
      </div>
    </main>
  );
}
