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
    <main className="pt-24 px-4 sm:px-6 lg:px-8 pb-16">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8 flex items-start justify-between">
          <div>
            <h1 className="text-4xl font-light tracking-tight mb-2">Overview</h1>
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

        {loading ? (
          <div className="flex justify-center py-12">
            <Loader2 className="h-6 w-6 text-white/60 animate-spin" />
          </div>
        ) : (
          <div className="space-y-6">
            {/* Key Metrics Cards */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
              <Link href="/admin/users" className="border border-white/10 rounded-lg p-6 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm hover:from-white/10 hover:to-white/5 transition-all">
                <div className="flex items-center gap-3 mb-2">
                  <Users className="h-5 w-5 text-white/60" />
                  <h3 className="text-lg font-light">Total Users</h3>
                </div>
                <p className="text-3xl font-light">{userCount}</p>
              </Link>
              <Link href="/admin/pis" className="border border-white/10 rounded-lg p-6 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm hover:from-white/10 hover:to-white/5 transition-all">
                <div className="flex items-center gap-3 mb-2">
                  <Server className="h-5 w-5 text-white/60" />
                  <h3 className="text-lg font-light">Total PIs</h3>
                </div>
                <p className="text-3xl font-light">{piCount}</p>
              </Link>
              <Link href="/admin/devices" className="border border-white/10 rounded-lg p-6 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm hover:from-white/10 hover:to-white/5 transition-all">
                <div className="flex items-center gap-3 mb-2">
                  <Cpu className="h-5 w-5 text-white/60" />
                  <h3 className="text-lg font-light">Total Devices</h3>
                </div>
                <p className="text-3xl font-light">{deviceCount}</p>
              </Link>
              {stats && stats.total_readings !== undefined && (
                <Link href="/admin/readings" className="border border-white/10 rounded-lg p-6 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm hover:from-white/10 hover:to-white/5 transition-all">
                  <div className="flex items-center gap-3 mb-2">
                    <BarChart3 className="h-5 w-5 text-white/60" />
                    <h3 className="text-lg font-light">Total Readings</h3>
                  </div>
                  <p className="text-3xl font-light">{(stats.total_readings || 0).toLocaleString()}</p>
                </Link>
              )}
            </div>

            {/* Health Status Cards */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="border border-white/10 rounded-lg p-6 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-lg font-light flex items-center gap-2">
                    <Activity className="h-5 w-5 text-white/60" />
                    API Service
                  </h3>
                  {healthLoading ? (
                    <Loader2 className="h-4 w-4 text-white/60 animate-spin" />
                  ) : apiHealth?.status === "ready" ? (
                    <CheckCircle2 className="h-5 w-5 text-green-400" />
                  ) : (
                    <XCircle className="h-5 w-5 text-red-400" />
                  )}
                </div>
                {apiHealth ? (
                  <div className="space-y-2">
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-white/60 font-light">Database</span>
                      <span className={apiHealth.db ? "text-green-400" : "text-red-400"}>
                        {apiHealth.db ? "Connected" : "Disconnected"}
                      </span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-white/60 font-light">MQTT</span>
                      <span className={apiHealth.mqtt ? "text-green-400" : "text-red-400"}>
                        {apiHealth.mqtt ? "Connected" : "Disconnected"}
                      </span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-white/60 font-light">Status</span>
                      <span className="text-white/80 font-light capitalize">{apiHealth.status}</span>
                    </div>
                  </div>
                ) : (
                  <p className="text-white/40 text-sm font-light">Unable to fetch health status</p>
                )}
              </div>

              <div className="border border-white/10 rounded-lg p-6 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-lg font-light flex items-center gap-2">
                    <Database className="h-5 w-5 text-white/60" />
                    Database
                  </h3>
                  {healthLoading ? (
                    <Loader2 className="h-4 w-4 text-white/60 animate-spin" />
                  ) : apiHealth?.db ? (
                    <CheckCircle2 className="h-5 w-5 text-green-400" />
                  ) : (
                    <XCircle className="h-5 w-5 text-red-400" />
                  )}
                </div>
                {apiHealth ? (
                  <div className="space-y-2">
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-white/60 font-light">Type</span>
                      <span className="text-white/80 font-light">PostgreSQL</span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-white/60 font-light">Status</span>
                      <span className={apiHealth.db ? "text-green-400" : "text-red-400"}>
                        {apiHealth.db ? "Connected" : "Disconnected"}
                      </span>
                    </div>
                  </div>
                ) : (
                  <p className="text-white/40 text-sm font-light">Unable to fetch database status</p>
                )}
              </div>
            </div>

            {/* Charts Section */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {readingsByPi.length > 0 && (
                <div className="border border-white/10 rounded-lg p-6 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm">
                  <h3 className="text-lg font-light mb-4">Readings Distribution by PI</h3>
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
                <div className="border border-white/10 rounded-lg p-6 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm lg:col-span-2">
                  <h3 className="text-lg font-light mb-4">Devices per PI</h3>
                  <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                    {devicesByPi
                      .sort((a, b) => b.devices - a.devices)
                      .map((pi, index) => (
                        <div
                          key={pi.name}
                          className="border border-white/10 rounded-lg p-4 bg-gradient-to-br from-white/5 to-white/0 hover:from-white/10 hover:to-white/5 transition-all"
                        >
                          <div className="flex items-center justify-between mb-2">
                            <div className="flex items-center gap-2">
                              <div
                                className="w-3 h-3 rounded-full"
                                style={{
                                  backgroundColor: ["#ea580c", "#f97316", "#fb923c", "#fdba74", "#fecaca", "#fef3c7"][index % 6],
                                }}
                              />
                              <span className="text-sm font-light text-white/80 truncate" title={pi.name}>
                                {pi.name}
                              </span>
                            </div>
                          </div>
                          <div className="mt-3">
                            <div className="text-2xl font-light">{pi.devices}</div>
                            <div className="text-xs text-white/50 font-light mt-1">
                              {pi.devices === 1 ? "device" : "devices"}
                            </div>
                          </div>
                          <div className="mt-3 h-2 bg-white/5 rounded-full overflow-hidden">
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

            {/* Additional Statistics */}
            {stats &&
              (stats.total_readings !== undefined ||
                stats.avg_humidity !== undefined ||
                (stats.min_temperature !== undefined && stats.max_temperature !== undefined)) && (
                <div className="border border-white/10 rounded-lg p-6 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm">
                  <h3 className="text-lg font-light mb-4">Sensor Statistics</h3>
                  <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                    {stats.total_readings !== undefined && (
                      <div className="border border-white/10 rounded-lg p-4 bg-white/5">
                        <div className="flex items-center gap-2 mb-2">
                          <BarChart3 className="h-4 w-4 text-white/60" />
                          <p className="text-white/60 text-sm font-light">Total Readings</p>
                        </div>
                        <p className="text-3xl font-light">{(stats.total_readings || 0).toLocaleString()}</p>
                        <p className="text-xs text-white/40 font-light mt-1">All sensor readings collected</p>
                      </div>
                    )}
                    {stats.avg_humidity !== undefined && (
                      <div className="border border-white/10 rounded-lg p-4 bg-white/5">
                        <div className="flex items-center gap-2 mb-2">
                          <Droplets className="h-4 w-4 text-white/60" />
                          <p className="text-white/60 text-sm font-light">Average Humidity</p>
                        </div>
                        <p className="text-3xl font-light">
                          {stats.avg_humidity.toFixed(1)}
                          <span className="text-lg text-white/60">%</span>
                        </p>
                        <p className="text-xs text-white/40 font-light mt-1">Across all sensors</p>
                      </div>
                    )}
                    {stats.min_temperature !== undefined && stats.max_temperature !== undefined && (
                      <div className="border border-white/10 rounded-lg p-4 bg-white/5">
                        <div className="flex items-center gap-2 mb-2">
                          <Thermometer className="h-4 w-4 text-white/60" />
                          <p className="text-white/60 text-sm font-light">Temperature Range</p>
                        </div>
                        <p className="text-2xl font-light">
                          {stats.min_temperature.toFixed(1)}° - {stats.max_temperature.toFixed(1)}°
                        </p>
                        <p className="text-xs text-white/40 font-light mt-1">Min to max recorded</p>
                      </div>
                    )}
                  </div>
                </div>
              )}
          </div>
        )}
      </div>
    </main>
  );
}
