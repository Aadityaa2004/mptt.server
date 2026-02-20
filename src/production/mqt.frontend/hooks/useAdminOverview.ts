"use client";

import { useState, useEffect, useCallback } from "react";
import { adminService } from "@/services/api/adminService";
import type { SummaryStatistics, PaginatedResponse, Device } from "@/types/admin";

export function useAdminOverview() {
  const [stats, setStats] = useState<SummaryStatistics | null>(null);
  const [userCount, setUserCount] = useState(0);
  const [piCount, setPiCount] = useState(0);
  const [deviceCount, setDeviceCount] = useState(0);
  const [readingsByPi, setReadingsByPi] = useState<{ name: string; value: number }[]>([]);
  const [devicesByPi, setDevicesByPi] = useState<{ name: string; devices: number }[]>([]);
  const [usersByRole, setUsersByRole] = useState<{ name: string; value: number }[]>([]);
  const [apiHealth, setApiHealth] = useState<{ db: boolean; mqtt: boolean; status: string } | null>(null);
  const [healthLoading, setHealthLoading] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadHealthChecks = useCallback(async () => {
    try {
      setHealthLoading(true);
      const apiData = await adminService.getApiHealth();
      setApiHealth(apiData);
    } catch (err) {
      console.error("Error loading health checks:", err);
      setApiHealth(null);
    } finally {
      setHealthLoading(false);
    }
  }, []);

  const loadOverviewData = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      const usersData = await adminService.getAllUsers().catch(() => ({ users: [] }));
      setUserCount(usersData?.users?.length || 0);

      const adminCount = usersData?.users?.filter((u) => u.role === "admin").length || 0;
      const regularUserCount = usersData?.users?.filter((u) => u.role === "user").length || 0;
      setUsersByRole([
        { name: "Admins", value: adminCount },
        { name: "Users", value: regularUserCount },
      ]);

      const pisData = await adminService.getAllPis(undefined, 1, 1000).catch(() => ({ items: [], total: 0 }));
      const piCountVal = pisData?.total ?? pisData?.items?.length ?? 0;
      setPiCount(piCountVal);

      const statsData = await adminService.getSummaryStats().catch(() => null);
      setStats(statsData);

      let totalDevices = 0;
      const devicesByPiData: { name: string; devices: number }[] = [];
      const readingsByPiData: { name: string; value: number }[] = [];

      if (pisData?.items && pisData.items.length > 0) {
        const piDataPromises = pisData.items.map(async (pi) => {
          try {
            const devicesData = await adminService.getDevices(pi.pi_id, 1, 1000) as PaginatedResponse<Device> & { next_page?: number | null };
            let dc = 0;
            if (devicesData?.total !== undefined) {
              dc = devicesData.total;
            } else if (devicesData?.items) {
              dc = devicesData.items.length;
              let currentPage = 1;
              let hasNextPage = devicesData.next_page != null;
              while (hasNextPage) {
                currentPage++;
                try {
                  const nextPageData = await adminService.getDevices(pi.pi_id, currentPage, 1000) as PaginatedResponse<Device> & { next_page?: number | null };
                  if (nextPageData?.items) {
                    dc += nextPageData.items.length;
                    hasNextPage = nextPageData.next_page != null;
                  } else hasNextPage = false;
                } catch {
                  hasNextPage = false;
                }
              }
            }
            let readingsCount = 0;
            try {
              const readingsData = await adminService.getReadings({ pi_id: pi.pi_id, page: 1, page_size: 1 });
              readingsCount = readingsData?.total || 0;
            } catch {
              readingsCount = 0;
            }
            return { piId: pi.pi_id, deviceCount: dc, readingsCount };
          } catch (err) {
            console.error(`Error loading data for PI ${pi.pi_id}:`, err);
            return { piId: pi.pi_id, deviceCount: 0, readingsCount: 0 };
          }
        });
        const results = await Promise.all(piDataPromises);
        totalDevices = results.reduce((sum, pi) => sum + pi.deviceCount, 0);
        results.forEach((pi) => {
          if (pi.deviceCount > 0) {
            devicesByPiData.push({
              name: pi.piId.length > 12 ? `${pi.piId.substring(0, 12)}...` : pi.piId,
              devices: pi.deviceCount,
            });
          }
          if (pi.readingsCount > 0) {
            readingsByPiData.push({
              name: pi.piId.length > 12 ? `${pi.piId.substring(0, 12)}...` : pi.piId,
              value: pi.readingsCount,
            });
          }
        });
      }
      setDeviceCount(totalDevices);
      setDevicesByPi(devicesByPiData);
      setReadingsByPi(readingsByPiData);
      await loadHealthChecks();
    } catch (err) {
      console.error("Error loading overview data:", err);
      setError(err instanceof Error ? err.message : "Failed to load dashboard data");
    } finally {
      setLoading(false);
    }
  }, [loadHealthChecks]);

  useEffect(() => {
    loadOverviewData();
  }, [loadOverviewData]);

  return {
    stats,
    userCount,
    piCount,
    deviceCount,
    readingsByPi,
    devicesByPi,
    usersByRole,
    apiHealth,
    healthLoading,
    loading,
    error,
    setError,
    loadOverviewData,
    loadHealthChecks,
  };
}
