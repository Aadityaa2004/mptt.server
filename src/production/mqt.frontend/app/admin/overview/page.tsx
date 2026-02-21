"use client";

import {
  Loader2,
  AlertCircle,
  Users,
  Server,
  Cpu,
  BarChart3,
  Activity,
  Database,
  CheckCircle2,
  XCircle,
  Thermometer,
  Droplets,
  RefreshCw,
} from "lucide-react";
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from "recharts";
import { useAdminOverview } from "@/hooks/useAdminOverview";
import Link from "next/link";

export default function AdminOverviewPage() {
  const {
    stats,
    userCount,
    piCount,
    deviceCount,
    readingsByPi,
    devicesByPi,
    apiHealth,
    healthLoading,
    loading,
    error,
    setError,
    loadOverviewData,
  } = useAdminOverview();

  const handleRefresh = async () => {
    setError(null);
    await loadOverviewData();
  };

  return (
    <main className="pt-16 sm:pt-20">
      <div className="max-w-[1600px] mx-auto">
        {/* Header */}
        <div className="px-4 sm:px-6 lg:px-8 py-6 border-b border-white/5 flex items-center justify-between">
          <h1 className="text-2xl sm:text-3xl font-light tracking-tight text-foreground">Overview</h1>
          <button
            onClick={handleRefresh}
            disabled={loading}
            className="p-2.5 rounded-xl bg-white/5 hover:bg-white/10 text-white/70 hover:text-white disabled:opacity-50 transition-all"
            title="Refresh data"
          >
            <RefreshCw className={`h-5 w-5 ${loading ? "animate-spin" : ""}`} />
          </button>
        </div>

        {error && (
          <div className="mx-4 sm:mx-6 lg:mx-8 mt-4 p-4 rounded-xl bg-red-500/5 border border-red-500/20 flex items-center gap-2">
            <AlertCircle className="h-5 w-5 text-red-400 flex-shrink-0" />
            <p className="text-sm text-red-400 font-light">{error}</p>
          </div>
        )}

        {loading ? (
          <div className="flex justify-center py-16">
            <Loader2 className="h-6 w-6 text-white/40 animate-spin" />
          </div>
        ) : (
          <div>
            {/* Key Metrics - compact strip */}
            <section className="px-4 sm:px-6 lg:px-8 py-6 border-b border-white/5">
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <Link href="/admin/users" className="flex items-center gap-4 p-4 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] border border-white/5 transition-all">
                  <div className="w-10 h-10 rounded-xl bg-orange-500/20 flex items-center justify-center">
                    <Users className="h-5 w-5 text-orange-400/90" />
                  </div>
                  <div>
                    <p className="text-2xl font-light">{userCount}</p>
                    <p className="text-xs text-white/50 font-light">Users</p>
                  </div>
                </Link>
                <Link href="/admin/pis" className="flex items-center gap-4 p-4 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] border border-white/5 transition-all">
                  <div className="w-10 h-10 rounded-xl bg-orange-500/20 flex items-center justify-center">
                    <Server className="h-5 w-5 text-orange-400/90" />
                  </div>
                  <div>
                    <p className="text-2xl font-light">{piCount}</p>
                    <p className="text-xs text-white/50 font-light">PIs</p>
                  </div>
                </Link>
                <Link href="/admin/devices" className="flex items-center gap-4 p-4 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] border border-white/5 transition-all">
                  <div className="w-10 h-10 rounded-xl bg-orange-500/20 flex items-center justify-center">
                    <Cpu className="h-5 w-5 text-orange-400/90" />
                  </div>
                  <div>
                    <p className="text-2xl font-light">{deviceCount}</p>
                    <p className="text-xs text-white/50 font-light">Devices</p>
                  </div>
                </Link>
                {stats && stats.total_readings !== undefined && (
                  <Link href="/admin/readings" className="flex items-center gap-4 p-4 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] border border-white/5 transition-all">
                    <div className="w-10 h-10 rounded-xl bg-orange-500/20 flex items-center justify-center">
                      <BarChart3 className="h-5 w-5 text-orange-400/90" />
                    </div>
                    <div>
                      <p className="text-2xl font-light">{(stats.total_readings || 0).toLocaleString()}</p>
                      <p className="text-xs text-white/50 font-light">Readings</p>
                    </div>
                  </Link>
                )}
              </div>
            </section>

            {/* Health Status - compact grid */}
            <section className="px-4 sm:px-6 lg:px-8 py-6 border-b border-white/5">
              <h2 className="text-lg font-light mb-4 text-white/90">System Health</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="p-5 rounded-xl bg-white/[0.04] border border-white/5">
                  <div className="flex items-center justify-between mb-3">
                    <span className="text-sm font-light text-white/80 flex items-center gap-2">
                      <Activity className="h-4 w-4 text-white/50" />
                      API Service
                    </span>
                    {healthLoading ? (
                      <Loader2 className="h-4 w-4 text-white/40 animate-spin" />
                    ) : apiHealth?.status === "ready" ? (
                      <CheckCircle2 className="h-5 w-5 text-green-400" />
                    ) : (
                      <XCircle className="h-5 w-5 text-red-400" />
                    )}
                  </div>
                  {apiHealth ? (
                    <div className="space-y-1.5 text-sm">
                      <div className="flex justify-between text-white/60">
                        <span>Database</span>
                        <span className={apiHealth.db ? "text-green-400" : "text-red-400"}>{apiHealth.db ? "Connected" : "Disconnected"}</span>
                      </div>
                      <div className="flex justify-between text-white/60">
                        <span>MQTT</span>
                        <span className={apiHealth.mqtt ? "text-green-400" : "text-red-400"}>{apiHealth.mqtt ? "Connected" : "Disconnected"}</span>
                      </div>
                    </div>
                  ) : (
                    <p className="text-white/40 text-sm font-light">Unable to fetch</p>
                  )}
                </div>
                <div className="p-5 rounded-xl bg-white/[0.04] border border-white/5">
                  <div className="flex items-center justify-between mb-3">
                    <span className="text-sm font-light text-white/80 flex items-center gap-2">
                      <Database className="h-4 w-4 text-white/50" />
                      Database
                    </span>
                    {healthLoading ? (
                      <Loader2 className="h-4 w-4 text-white/40 animate-spin" />
                    ) : apiHealth?.db ? (
                      <CheckCircle2 className="h-5 w-5 text-green-400" />
                    ) : (
                      <XCircle className="h-5 w-5 text-red-400" />
                    )}
                  </div>
                  {apiHealth ? (
                    <div className="space-y-1.5 text-sm">
                      <div className="flex justify-between text-white/60">
                        <span>Type</span>
                        <span>PostgreSQL</span>
                      </div>
                      <div className="flex justify-between text-white/60">
                        <span>Status</span>
                        <span className={apiHealth.db ? "text-green-400" : "text-red-400"}>{apiHealth.db ? "Connected" : "Disconnected"}</span>
                      </div>
                    </div>
                  ) : (
                    <p className="text-white/40 text-sm font-light">Unable to fetch</p>
                  )}
                </div>
              </div>
            </section>

            {/* Charts Section */}
            <section className="px-4 sm:px-6 lg:px-8 py-6">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {readingsByPi.length > 0 && (
                <div className="p-6 rounded-2xl bg-white/[0.04] border border-white/5">
                  <h3 className="text-lg font-light mb-4 text-white/90">Readings by PI</h3>
                  <ResponsiveContainer width="100%" height={300}>
                    <PieChart>
                      <Pie
                        data={readingsByPi}
                        cx="50%"
                        cy="50%"
                        labelLine={false}
                        label={({ name, percent }) => `${name}: ${(percent * 100).toFixed(0)}%`}
                        outerRadius={80}
                        fill="#8884d8"
                        dataKey="value"
                      >
                        {readingsByPi.map((entry, index) => {
                          const colors = ["#ea580c", "#f97316", "#fb923c", "#fdba74", "#fecaca", "#fef3c7"];
                          return <Cell key={`cell-${index}`} fill={colors[index % colors.length]} />;
                        })}
                      </Pie>
                      <Tooltip
                        formatter={(value: number) => value.toLocaleString()}
                        contentStyle={{ backgroundColor: "rgba(0, 0, 0, 0.8)", border: "1px solid rgba(255, 255, 255, 0.1)", borderRadius: "8px" }}
                      />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
              )}

              {devicesByPi.length > 0 && (
                <div className="p-6 rounded-2xl bg-white/[0.04] border border-white/5 lg:col-span-2">
                  <h3 className="text-lg font-light mb-4 text-white/90">Devices per PI</h3>
                  <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
                    {devicesByPi
                      .sort((a, b) => b.devices - a.devices)
                      .map((pi, index) => (
                        <div
                          key={pi.name}
                          className="p-4 rounded-xl bg-white/[0.04] border border-white/5 hover:border-white/10 transition-all"
                        >
                          <div className="flex items-center gap-2 mb-2">
                            <div
                              className="w-2.5 h-2.5 rounded-full flex-shrink-0"
                              style={{
                                backgroundColor: ["#ea580c", "#f97316", "#fb923c", "#fdba74", "#fecaca", "#fef3c7"][index % 6],
                              }}
                            />
                            <span className="text-sm font-light text-white/90 truncate" title={pi.name}>
                              {pi.name}
                            </span>
                          </div>
                          <div className="flex items-baseline gap-1">
                            <span className="text-xl font-light">{pi.devices}</span>
                            <span className="text-xs text-white/50 font-light">{pi.devices === 1 ? "device" : "devices"}</span>
                          </div>
                          <div className="mt-2 h-1.5 bg-white/5 rounded-full overflow-hidden">
                            <div
                              className="h-full rounded-full transition-all"
                              style={{
                                width: `${(pi.devices / Math.max(...devicesByPi.map((p) => p.devices))) * 100}%`,
                                backgroundColor: ["#ea580c", "#f97316", "#fb923c", "#fdba74", "#fecaca", "#fef3c7"][index % 6],
                              }}
                            />
                          </div>
                        </div>
                      ))}
                  </div>
                </div>
              )}
            </div>

            {/* Sensor Statistics */}
            {stats &&
              (stats.total_readings !== undefined ||
                stats.avg_humidity !== undefined ||
                (stats.min_temperature !== undefined && stats.max_temperature !== undefined)) && (
                <div className="mt-6 p-6 rounded-2xl bg-white/[0.04] border border-white/5">
                  <h3 className="text-lg font-light mb-4 text-white/90">Sensor Statistics</h3>
                  <div className="flex flex-wrap gap-6">
                    {stats.total_readings !== undefined && (
                      <div className="flex items-center gap-4">
                        <BarChart3 className="h-5 w-5 text-orange-400/80" />
                        <div>
                          <p className="text-2xl font-light">{(stats.total_readings || 0).toLocaleString()}</p>
                          <p className="text-xs text-white/50 font-light">Total readings</p>
                        </div>
                      </div>
                    )}
                    {stats.avg_humidity !== undefined && (
                      <div className="flex items-center gap-4">
                        <Droplets className="h-5 w-5 text-orange-400/80" />
                        <div>
                          <p className="text-2xl font-light">{stats.avg_humidity.toFixed(1)}%</p>
                          <p className="text-xs text-white/50 font-light">Avg humidity</p>
                        </div>
                      </div>
                    )}
                    {stats.min_temperature !== undefined && stats.max_temperature !== undefined && (
                      <div className="flex items-center gap-4">
                        <Thermometer className="h-5 w-5 text-orange-400/80" />
                        <div>
                          <p className="text-2xl font-light">{stats.min_temperature.toFixed(1)}° – {stats.max_temperature.toFixed(1)}°</p>
                          <p className="text-xs text-white/50 font-light">Temp range</p>
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              )}
            </section>
          </div>
        )}
      </div>
    </main>
  );
}
