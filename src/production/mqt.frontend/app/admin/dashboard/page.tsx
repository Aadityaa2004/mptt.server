"use client";

import { useState, useEffect } from "react";
import { useRequireAuth } from "@/hooks/useRequireAuth";
import Navbar from "@/components/navbar/Navbar";
import { adminService } from "@/services/api/adminService";
import { ReadingsChart } from "@/components/sensors/ReadingsChart";
import { Loader2, AlertCircle, Users, Server, Cpu, BarChart3, X, ArrowLeft, Thermometer, Droplets, Battery, RefreshCw, Activity, Database, CheckCircle2, XCircle, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { PieChart, Pie, Cell, ResponsiveContainer, BarChart, Bar, XAxis, YAxis, Tooltip, Legend, CartesianGrid } from "recharts";
import type { User, Pi, Device, Reading, SummaryStatistics, PaginatedResponse } from "@/types/admin";

type Tab = "overview" | "users" | "pis" | "devices" | "readings";

export default function AdminDashboardPage() {
  const { user, isLoading } = useRequireAuth("admin");
  const [activeTab, setActiveTab] = useState<Tab>("overview");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  // Overview data
  const [stats, setStats] = useState<SummaryStatistics | null>(null);
  const [userCount, setUserCount] = useState(0);
  const [piCount, setPiCount] = useState(0);
  const [deviceCount, setDeviceCount] = useState(0);
  const [readingsByPi, setReadingsByPi] = useState<{ name: string; value: number }[]>([]);
  const [devicesByPi, setDevicesByPi] = useState<{ name: string; devices: number }[]>([]);
  const [usersByRole, setUsersByRole] = useState<{ name: string; value: number }[]>([]);
  
  // Health check data
  const [apiHealth, setApiHealth] = useState<{
    db: boolean;
    mqtt: boolean;
    status: string;
  } | null>(null);
  const [healthLoading, setHealthLoading] = useState(false);

  // Users data
  const [users, setUsers] = useState<User[]>([]);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [showUserForm, setShowUserForm] = useState(false);
  const [showDeleteUserConfirm, setShowDeleteUserConfirm] = useState<string | null>(null);

  // PIs data
  const [pis, setPis] = useState<Pi[]>([]);
  const [selectedPi, setSelectedPi] = useState<Pi | null>(null);
  const [showPiForm, setShowPiForm] = useState(false);
  const [showDeletePiConfirm, setShowDeletePiConfirm] = useState<string | null>(null);
  const [piDevices, setPiDevices] = useState<Device[]>([]);

  // Devices data
  const [devices, setDevices] = useState<Device[]>([]);
  const [allDevices, setAllDevices] = useState<Device[]>([]); // All devices from all PIs
  const [selectedPiFilter, setSelectedPiFilter] = useState<string>(""); // Filter for devices view
  const [selectedDevice, setSelectedDevice] = useState<{ piId: string; deviceId: number } | null>(null);
  const [showDeviceForm, setShowDeviceForm] = useState(false);
  const [showDeleteDeviceConfirm, setShowDeleteDeviceConfirm] = useState<{ piId: string; deviceId: number | string } | null>(null);

  // Readings data
  const [readings, setReadings] = useState<Reading[]>([]);
  const [selectedPiForReadings, setSelectedPiForReadings] = useState<string>("");
  const [selectedDeviceForReadings, setSelectedDeviceForReadings] = useState<string | null>(null);
  const [showDeviceAnalytics, setShowDeviceAnalytics] = useState(false);
  const [latestReading, setLatestReading] = useState<Reading | null>(null);
  const [readingsForChart, setReadingsForChart] = useState<Reading[]>([]);
  const [isLoadingReadings, setIsLoadingReadings] = useState(false);
  const [timeRange, setTimeRange] = useState<"1h" | "1d" | "1w" | "1m" | "1y">("1d");
  const [isRefreshing, setIsRefreshing] = useState(false);

  // Form states
  const [userFormData, setUserFormData] = useState({
    username: "",
    email: "",
    password: "",
    role: "user" as "admin" | "user",
  });
  const [piFormData, setPiFormData] = useState({
    pi_id: "",
    user_id: "",
  });
  const [deviceFormData, setDeviceFormData] = useState({
    device_id: "",
  });

  useEffect(() => {
    if (!isLoading && user) {
      loadOverviewData();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isLoading, user]);

  useEffect(() => {
    if (activeTab === "overview") {
      loadOverviewData();
    } else if (activeTab === "users") {
      loadUsers();
    } else if (activeTab === "pis") {
      loadPis();
      // Load users for assignment dropdown
      if (users.length === 0) {
        adminService.getAllUsers()
          .then((data) => setUsers(data?.users || []))
          .catch((err) => {
            console.error("Error loading users for PI assignment:", err);
            setUsers([]);
          });
      }
    } else if (activeTab === "devices") {
      if (selectedPi) {
        loadDevices(selectedPi.pi_id);
      } else {
        loadAllDevices();
        // Load PIs for the filter dropdown if not already loaded
        if (pis.length === 0) {
          loadPis();
        }
      }
    } else if (activeTab === "readings") {
      // Load PIs for the dropdown
      if (pis.length === 0) {
        loadPis();
      }
      if (selectedPiForReadings) {
        // Load devices for the selected PI
        adminService.getDevices(selectedPiForReadings)
          .then((data) => setDevices(data?.items || []))
          .catch((err) => {
            console.error("Error loading devices for readings:", err);
            setDevices([]);
          });
        if (!showDeviceAnalytics) {
        loadReadings();
        }
      }
      if (showDeviceAnalytics && selectedPiForReadings && selectedDeviceForReadings) {
        loadDeviceAnalytics();
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activeTab, selectedPi, selectedPiForReadings, selectedDeviceForReadings, showDeviceAnalytics]);

  // Clear confirmation dialogs when navigating away from relevant views
  useEffect(() => {
    if (activeTab !== "devices") {
      setShowDeleteDeviceConfirm(null);
    }
    if (activeTab !== "users") {
      setShowDeleteUserConfirm(null);
    }
    if (activeTab !== "pis") {
      setShowDeletePiConfirm(null);
    }
  }, [activeTab]);

  // Clear device confirmation dialog when selectedPi changes or when devices are reloaded
  useEffect(() => {
    // Clear confirmation if the device being confirmed for deletion no longer exists
    if (showDeleteDeviceConfirm) {
      const { piId, deviceId } = showDeleteDeviceConfirm;
      if (selectedPi && selectedPi.pi_id === piId) {
        // Check if device still exists in piDevices
        const deviceExists = piDevices.some(
          (d) => String(d.device_id) === String(deviceId) && d.pi_id === piId
        );
        if (!deviceExists) {
          setShowDeleteDeviceConfirm(null);
        }
      } else if (!selectedPi) {
        // Check if device still exists in allDevices
        const deviceExists = allDevices.some(
          (d) => String(d.device_id) === String(deviceId) && d.pi_id === piId
        );
        if (!deviceExists) {
          setShowDeleteDeviceConfirm(null);
        }
      }
    }
  }, [showDeleteDeviceConfirm, selectedPi, piDevices, allDevices]);

  const loadHealthChecks = async () => {
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
  };

  const loadOverviewData = async () => {
    try {
      setLoading(true);
      setError(null);
      
      // Fetch all users
      const usersData = await adminService.getAllUsers().catch(() => ({ users: [] }));
      setUserCount(usersData?.users?.length || 0);
      
      // Calculate users by role
      const adminCount = usersData?.users?.filter(u => u.role === "admin").length || 0;
      const regularUserCount = usersData?.users?.filter(u => u.role === "user").length || 0;
      setUsersByRole([
        { name: "Admins", value: adminCount },
        { name: "Users", value: regularUserCount },
      ]);
      
      // Fetch all PIs with a large page size to get accurate count
      const pisData = await adminService.getAllPis(undefined, 1, 1000).catch(() => ({ items: [], total: 0, page: 1, page_size: 1000 }));
      const piCount = pisData?.total || pisData?.items?.length || 0;
      setPiCount(piCount);
      
      // Fetch stats
      const statsData = await adminService.getSummaryStats().catch(() => null);
      setStats(statsData);
      
      // Calculate total device count and readings distribution by PI
      let totalDevices = 0;
      const devicesByPiData: { name: string; devices: number }[] = [];
      const readingsByPiData: { name: string; value: number }[] = [];
      
      if (pisData?.items && pisData.items.length > 0) {
        try {
          const piDataPromises = pisData.items.map(async (pi) => {
              try {
              // Fetch devices for this PI
                const devicesData = await adminService.getDevices(pi.pi_id, 1, 1000);
                const response = devicesData as PaginatedResponse<Device> & { next_page?: number | null };
              let deviceCount = 0;
              
                if (response?.total !== undefined) {
                deviceCount = response.total;
                } else if (response?.items) {
                deviceCount = response.items.length;
                  let currentPage = 1;
                  let hasNextPage = response.next_page !== undefined && response.next_page !== null;
                  
                  while (hasNextPage) {
                    currentPage++;
                    try {
                      const nextPageData = await adminService.getDevices(pi.pi_id, currentPage, 1000) as PaginatedResponse<Device> & { next_page?: number | null };
                      if (nextPageData?.items) {
                      deviceCount += nextPageData.items.length;
                        hasNextPage = nextPageData.next_page !== undefined && nextPageData.next_page !== null;
                      } else {
                        hasNextPage = false;
                      }
                    } catch {
                      hasNextPage = false;
                    }
                  }
              }
              
              // Fetch readings count for this PI
              let readingsCount = 0;
              try {
                const readingsData = await adminService.getReadings({ pi_id: pi.pi_id, page: 1, page_size: 1 });
                readingsCount = readingsData?.total || 0;
              } catch {
                // If total not available, try to estimate
                readingsCount = 0;
              }
              
              return {
                piId: pi.pi_id,
                deviceCount,
                readingsCount,
              };
              } catch (err) {
              console.error(`Error loading data for PI ${pi.pi_id}:`, err);
              return { piId: pi.pi_id, deviceCount: 0, readingsCount: 0 };
            }
          });
          
          const piDataResults = await Promise.all(piDataPromises);
          
          totalDevices = piDataResults.reduce((sum, pi) => sum + pi.deviceCount, 0);
          
          // Build charts data
          piDataResults.forEach((pi) => {
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
          
          setDevicesByPi(devicesByPiData);
          setReadingsByPi(readingsByPiData);
        } catch (err) {
          console.error("Error counting devices:", err);
          totalDevices = 0;
        }
      }
      setDeviceCount(totalDevices);
      
      // Load health checks
      await loadHealthChecks();
      
      console.log("Overview data loaded:", {
        users: usersData?.users?.length || 0,
        pis: piCount,
        devices: totalDevices,
        pisData: pisData
      });
    } catch (err) {
      console.error("Error loading overview data:", err);
      setError(err instanceof Error ? err.message : "Failed to load dashboard data");
    } finally {
      setLoading(false);
    }
  };

  const loadUsers = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await adminService.getAllUsers();
      setUsers(data?.users || []);
    } catch (err) {
      console.error("Error loading users:", err);
      setError(err instanceof Error ? err.message : "Failed to load users");
      setUsers([]);
    } finally {
      setLoading(false);
    }
  };

  const loadPis = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await adminService.getAllPis();
      setPis(data?.items || []);
    } catch (err) {
      console.error("Error loading PIs:", err);
      setError(err instanceof Error ? err.message : "Failed to load PIs");
      setPis([]);
    } finally {
      setLoading(false);
    }
  };

  const loadDevices = async (piId: string) => {
    try {
      setLoading(true);
      setError(null);
      const data = await adminService.getDevices(piId);
      const items = data?.items || [];
      setPiDevices(items);
      setDevices(items);
    } catch (err) {
      console.error("Error loading devices:", err);
      setError(err instanceof Error ? err.message : "Failed to load devices");
      setPiDevices([]);
      setDevices([]);
    } finally {
      setLoading(false);
    }
  };

  const loadAllDevices = async () => {
    try {
      setLoading(true);
      setError(null);
      
      // First load all PIs if not already loaded
      let pisToUse = pis;
      if (pisToUse.length === 0) {
        const pisData = await adminService.getAllPis(undefined, 1, 1000);
        pisToUse = pisData?.items || [];
      }
      
      // Fetch devices from all PIs
      const allDevicesData = await Promise.all(
        pisToUse.map(async (pi) => {
          try {
            const devicesData = await adminService.getDevices(pi.pi_id, 1, 1000);
            return devicesData?.items || [];
          } catch (err) {
            console.error(`Error loading devices for PI ${pi.pi_id}:`, err);
            return [];
          }
        })
      );
      
      // Flatten the array
      const flattenedDevices = allDevicesData.flat();
      setAllDevices(flattenedDevices);
    } catch (err) {
      console.error("Error loading all devices:", err);
      setError(err instanceof Error ? err.message : "Failed to load devices");
      setAllDevices([]);
    } finally {
      setLoading(false);
    }
  };

  const loadReadings = async () => {
    try {
      setLoading(true);
      setError(null);
      if (selectedDeviceForReadings !== null && selectedDeviceForReadings !== "") {
        // Convert device ID to number if it's numeric, otherwise use as string
        const deviceId = selectedDeviceForReadings;
        const isNumeric = /^\d+$/.test(deviceId);
        const deviceIdParam = isNumeric ? parseInt(deviceId, 10) : deviceId;
        const data = await adminService.getDeviceReadings(selectedPiForReadings, deviceIdParam);
        setReadings(data?.items || []);
      } else {
        const data = await adminService.getReadings({ pi_id: selectedPiForReadings });
        setReadings(data?.items || []);
      }
    } catch (err) {
      console.error("Error loading readings:", err);
      setError(err instanceof Error ? err.message : "Failed to load readings");
      setReadings([]);
    } finally {
      setLoading(false);
    }
  };

  const loadDeviceAnalytics = async () => {
    if (!selectedPiForReadings || !selectedDeviceForReadings) return;

    try {
      setIsLoadingReadings(true);
      setError(null);

      const deviceId = selectedDeviceForReadings;
      const isNumeric = /^\d+$/.test(deviceId);
      const deviceIdParam = isNumeric ? parseInt(deviceId, 10) : deviceId;

      // Load latest reading
      const latestResponse = await adminService.getDeviceReadings(selectedPiForReadings, deviceIdParam, {
        page: 1,
        page_size: 1,
      });
      if (latestResponse?.items && latestResponse.items.length > 0) {
        setLatestReading(latestResponse.items[0]);
      }

      // Load readings for chart
      const chartResponse = await adminService.getDeviceReadings(selectedPiForReadings, deviceIdParam, {
        page: 1,
        page_size: 100,
      });
      setReadingsForChart(chartResponse?.items || []);
    } catch (err) {
      console.error("Error loading device analytics:", err);
      setError(err instanceof Error ? err.message : "Failed to load device analytics");
      setReadingsForChart([]);
      setLatestReading(null);
    } finally {
      setIsLoadingReadings(false);
    }
  };

  const refreshDeviceAnalytics = async () => {
    if (!selectedPiForReadings || !selectedDeviceForReadings || isRefreshing) return;

    try {
      setIsRefreshing(true);
      setError(null);
      await loadDeviceAnalytics();
    } catch (err) {
      console.error("Error refreshing device analytics:", err);
      const errorMessage = err instanceof Error ? err.message : "Failed to refresh device analytics";
      setError(errorMessage);
    } finally {
      setIsRefreshing(false);
    }
  };

  const handleDeviceClick = async (deviceId: string) => {
    setSelectedDeviceForReadings(deviceId);
    setShowDeviceAnalytics(true);
  };

  const handleViewDeviceReadings = async (piId: string, deviceId: number | string) => {
    // Clear any confirmation dialogs
    setShowDeleteDeviceConfirm(null);
    // Ensure PIs are loaded
    if (pis.length === 0) {
      await loadPis();
    }
    // Set the readings tab state
    setSelectedPiForReadings(piId);
    setSelectedDeviceForReadings(String(deviceId));
    setShowDeviceAnalytics(true);
    // Switch to readings tab
    setActiveTab("readings");
    // Load devices for the selected PI
    try {
      const devicesData = await adminService.getDevices(piId);
      setDevices(devicesData?.items || []);
    } catch (err) {
      console.error("Error loading devices for readings:", err);
      setDevices([]);
    }
  };

  const handleBackToReadings = () => {
    setShowDeviceAnalytics(false);
    setSelectedDeviceForReadings(null);
    setLatestReading(null);
    setReadingsForChart([]);
  };

  const handleCreateUser = async () => {
    try {
      setLoading(true);
      setError(null);
      await adminService.registerAdmin(userFormData);
      setShowUserForm(false);
      setUserFormData({ username: "", email: "", password: "", role: "user" });
      await loadUsers();
      // Refresh overview statistics
      if (activeTab === "overview") {
        await loadOverviewData();
      }
    } catch (err) {
      console.error("Error creating user:", err);
      setError(err instanceof Error ? err.message : "Failed to create user");
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateUserRole = async (userId: string, role: "admin" | "user") => {
    try {
      setLoading(true);
      setError(null);
      await adminService.updateUserRole(userId, role);
      await loadUsers();
    } catch (err) {
      console.error("Error updating user role:", err);
      setError(err instanceof Error ? err.message : "Failed to update user role");
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteUser = async (userId: string) => {
    try {
      setLoading(true);
      setError(null);
      await adminService.deleteUser(userId);
      setShowDeleteUserConfirm(null);
      await loadUsers();
      // Refresh overview statistics
      if (activeTab === "overview") {
        await loadOverviewData();
      }
    } catch (err) {
      console.error("Error deleting user:", err);
      setError(err instanceof Error ? err.message : "Failed to delete user");
    } finally {
      setLoading(false);
    }
  };

  const handleCreatePi = async () => {
    try {
      setLoading(true);
      setError(null);
      // Ensure users are loaded
      if (users.length === 0) {
        const usersData = await adminService.getAllUsers();
        setUsers(usersData.users);
      }
      await adminService.createPi({
        pi_id: piFormData.pi_id,
        user_id: piFormData.user_id || undefined,
      });
      setShowPiForm(false);
      setPiFormData({ pi_id: "", user_id: "" });
      await loadPis();
      // Refresh overview statistics
      if (activeTab === "overview") {
        await loadOverviewData();
      }
    } catch (err) {
      console.error("Error creating PI:", err);
      setError(err instanceof Error ? err.message : "Failed to create PI");
    } finally {
      setLoading(false);
    }
  };

  const handleUpdatePi = async (piId: string, userId: string) => {
    try {
      setLoading(true);
      setError(null);
      // Ensure users are loaded
      if (users.length === 0) {
        const usersData = await adminService.getAllUsers();
        setUsers(usersData.users);
      }
      await adminService.updatePi(piId, { user_id: userId || undefined });
      await loadPis();
    } catch (err) {
      console.error("Error updating PI:", err);
      setError(err instanceof Error ? err.message : "Failed to update PI");
    } finally {
      setLoading(false);
    }
  };

  const handleDeletePi = async (piId: string) => {
    try {
      setLoading(true);
      setError(null);
      await adminService.deletePi(piId, true);
      setShowDeletePiConfirm(null);
      await loadPis();
      // Refresh overview statistics
      if (activeTab === "overview") {
        await loadOverviewData();
      }
    } catch (err) {
      console.error("Error deleting PI:", err);
      setError(err instanceof Error ? err.message : "Failed to delete PI");
    } finally {
      setLoading(false);
    }
  };

  // Validate MAC address format (e.g., AA:BB:CC:DD:EE:FF or AA-BB-CC-DD-EE-FF)
  const isValidMACAddress = (mac: string): boolean => {
    const macRegex = /^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$/;
    return macRegex.test(mac);
  };

  // Convert MAC address to a number (for backend compatibility)
  const macAddressToNumber = (mac: string): number => {
    // Remove colons and dashes, convert hex to decimal
    const hexString = mac.replace(/[:-]/g, "");
    // Convert hex to decimal, but limit to safe integer range
    // Use modulo to keep it within reasonable range
    const decimal = parseInt(hexString, 16);
    // Use modulo to keep it within JavaScript safe integer range, but still unique
    return decimal % Number.MAX_SAFE_INTEGER;
  };

  // Validate if device_id is either a number or text (MAC address format or any text)
  const isValidDeviceId = (id: string): boolean => {
    if (!id || !id.trim()) return false;
    const trimmedId = id.trim();
    // Check if it's a valid number
    if (!isNaN(Number(trimmedId)) && Number(trimmedId) > 0) return true;
    // Check if it's a valid MAC address format (with colons or dashes)
    if (isValidMACAddress(trimmedId)) return true;
    // Allow any non-empty text string (backend accepts text format)
    // Just ensure it has some content
    return trimmedId.length > 0;
  };

  const handleCreateDevice = async () => {
    if (!selectedPi) return;
    
    // Validate form data
    if (!deviceFormData.device_id || !isValidDeviceId(deviceFormData.device_id)) {
      setError("Device ID is required and must be a valid number or MAC address (e.g., AA:BB:CC:DD:EE:FF)");
      return;
    }
    
    try {
      setLoading(true);
      setError(null);
      
      // Always send device_id as a string (backend expects string type)
      // Can be numeric string (e.g., "123") or MAC address string (e.g., "AA:BB:CC:DD:EE:FF")
      const deviceId = deviceFormData.device_id.trim();
      
      await adminService.createDevice(selectedPi.pi_id, {
        device_id: deviceId,
      });
      setShowDeviceForm(false);
      setDeviceFormData({ 
        device_id: "", 
      });
      await loadDevices(selectedPi.pi_id);
      // Refresh overview statistics
      if (activeTab === "overview") {
        await loadOverviewData();
      }
    } catch (err) {
      console.error("Error creating device:", err);
      setError(err instanceof Error ? err.message : "Failed to create device");
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateDevice = async (piId: string, deviceId: number | string) => {
    try {
      setLoading(true);
      setError(null);
      await adminService.updateDevice(piId, deviceId, {});
      await loadDevices(piId);
    } catch (err) {
      console.error("Error updating device:", err);
      setError(err instanceof Error ? err.message : "Failed to update device");
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteDevice = async (piId: string, deviceId: number | string) => {
    if (!piId || deviceId === undefined || deviceId === null || deviceId === "") {
      setError("Invalid device ID or PI ID");
      setShowDeleteDeviceConfirm(null);
      return;
    }
    
    try {
      setLoading(true);
      setError(null);
      await adminService.deleteDevice(piId, deviceId, true);
      setShowDeleteDeviceConfirm(null);
      if (selectedPi) {
        await loadDevices(piId);
      } else {
        await loadAllDevices();
      }
      // Refresh overview statistics
      if (activeTab === "overview") {
        await loadOverviewData();
      }
    } catch (err) {
      console.error("Error deleting device:", err);
      setError(err instanceof Error ? err.message : "Failed to delete device");
      setShowDeleteDeviceConfirm(null);
    } finally {
      setLoading(false);
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

  const tabs: { id: Tab; label: string; icon: React.ReactNode }[] = [
    { id: "overview", label: "Overview", icon: <BarChart3 className="h-4 w-4" /> },
    { id: "users", label: "Users", icon: <Users className="h-4 w-4" /> },
    { id: "pis", label: "PIs", icon: <Server className="h-4 w-4" /> },
    { id: "devices", label: "Devices", icon: <Cpu className="h-4 w-4" /> },
    { id: "readings", label: "Readings", icon: <BarChart3 className="h-4 w-4" /> },
  ];

  return (
    <div className="min-h-screen bg-black text-white">
      <Navbar />
      <main className="pt-24 px-4 sm:px-6 lg:px-8 pb-16">
        <div className="max-w-7xl mx-auto">
          <div className="mb-8">
            <h1 className="text-4xl font-light tracking-tight mb-2">Dashboard</h1>
            {/* <p className="text-white/60 font-light text-sm">Manage users, PIs, devices, and system settings</p> */}
          </div>

          {error && (
            <div className="mb-6 p-4 border border-red-500/20 bg-red-500/10 rounded-lg flex items-center gap-2">
              <AlertCircle className="h-5 w-5 text-red-400" />
              <p className="text-sm text-red-400 font-light">{error}</p>
            </div>
          )}

          {/* Tabs */}
          <div className="border-b border-white/10 mb-6">
            <div className="flex gap-2 overflow-x-auto">
              {tabs.map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => {
                    setActiveTab(tab.id);
                    setError(null);
                    // Clear all confirmation dialogs when switching tabs
                    setShowDeleteUserConfirm(null);
                    setShowDeletePiConfirm(null);
                    setShowDeleteDeviceConfirm(null);
                  }}
                  className={`flex items-center gap-2 px-4 py-2 border-b-2 transition-colors font-light ${
                    activeTab === tab.id
                      ? "border-white text-white"
                      : "border-transparent text-white/60 hover:text-white/80"
                  }`}
                >
                  {tab.icon}
                  {tab.label}
                </button>
              ))}
            </div>
          </div>

          {/* Tab Content */}
          {loading && (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-6 w-6 text-white/60 animate-spin" />
            </div>
          )}

          {!loading && activeTab === "overview" && (
            <div className="space-y-6">
              {/* Key Metrics Cards */}
              <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
                <div className="border border-white/10 rounded-lg p-6 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm">
                <div className="flex items-center gap-3 mb-2">
                  <Users className="h-5 w-5 text-white/60" />
                  <h3 className="text-lg font-light">Total Users</h3>
                </div>
                <p className="text-3xl font-light">{userCount}</p>
              </div>
                <div className="border border-white/10 rounded-lg p-6 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm">
                <div className="flex items-center gap-3 mb-2">
                  <Server className="h-5 w-5 text-white/60" />
                  <h3 className="text-lg font-light">Total PIs</h3>
                </div>
                <p className="text-3xl font-light">{piCount}</p>
              </div>
                <div className="border border-white/10 rounded-lg p-6 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm">
                <div className="flex items-center gap-3 mb-2">
                  <Cpu className="h-5 w-5 text-white/60" />
                  <h3 className="text-lg font-light">Total Devices</h3>
                </div>
                <p className="text-3xl font-light">{deviceCount}</p>
              </div>
                {stats && stats.total_readings !== undefined && (
                  <div className="border border-white/10 rounded-lg p-6 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm">
                    <div className="flex items-center gap-3 mb-2">
                      <BarChart3 className="h-5 w-5 text-white/60" />
                      <h3 className="text-lg font-light">Total Readings</h3>
                    </div>
                    <p className="text-3xl font-light">{(stats.total_readings || 0).toLocaleString()}</p>
                  </div>
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
                {/* Readings Distribution Pie Chart */}
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

                {/* Devices by PI - Card Grid */}
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
                                    backgroundColor: ["#ea580c", "#f97316", "#fb923c", "#fdba74", "#fecaca", "#fef3c7"][index % 6]
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
                                  width: `${(pi.devices / Math.max(...devicesByPi.map(p => p.devices))) * 100}%`,
                                  backgroundColor: ["#ea580c", "#f97316", "#fb923c", "#fdba74", "#fecaca", "#fef3c7"][index % 6]
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
              {stats && (stats.total_readings !== undefined || stats.avg_humidity !== undefined || (stats.min_temperature !== undefined && stats.max_temperature !== undefined)) && (
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
                        <p className="text-3xl font-light">{stats.avg_humidity.toFixed(1)}<span className="text-lg text-white/60">%</span></p>
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

          {!loading && activeTab === "users" && (
            <div>
              <div className="flex justify-between items-center mb-6">
                <h2 className="text-2xl font-light">User Management</h2>
                <Button onClick={() => setShowUserForm(true)}>Create User</Button>
              </div>

              {showUserForm && (
                <div className="border border-white/10 rounded-lg p-6 bg-black/50 mb-6">
                  <div className="flex justify-between items-center mb-4">
                    <h3 className="text-lg font-light">Create New User</h3>
                    <button onClick={() => setShowUserForm(false)} className="text-white/60 hover:text-white">
                      <X className="h-5 w-5" />
                    </button>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <Input
                      placeholder="Username"
                      value={userFormData.username}
                      onChange={(e) => setUserFormData({ ...userFormData, username: e.target.value })}
                      className="bg-black/50 border-white/10"
                    />
                    <Input
                      type="email"
                      placeholder="Email"
                      value={userFormData.email}
                      onChange={(e) => setUserFormData({ ...userFormData, email: e.target.value })}
                      className="bg-black/50 border-white/10"
                    />
                    <Input
                      type="password"
                      placeholder="Password"
                      value={userFormData.password}
                      onChange={(e) => setUserFormData({ ...userFormData, password: e.target.value })}
                      className="bg-black/50 border-white/10"
                    />
                    <select
                      value={userFormData.role}
                      onChange={(e) => setUserFormData({ ...userFormData, role: e.target.value as "admin" | "user" })}
                      className="flex h-9 w-full rounded-md border border-white bg-black px-3 py-1 text-sm"
                    >
                      <option value="user">User</option>
                      <option value="admin">Admin</option>
                    </select>
                  </div>
                  <div className="flex gap-2 mt-4">
                    <Button onClick={handleCreateUser}>Create</Button>
                    <Button variant="outline" onClick={() => setShowUserForm(false)}>Cancel</Button>
                  </div>
                </div>
              )}

              <div className="border border-white/10 rounded-lg overflow-hidden">
                <table className="w-full">
                  <thead className="bg-black/50 border-b border-white/10">
                    <tr>
                      <th className="px-4 py-3 text-left text-sm font-light">Username</th>
                      <th className="px-4 py-3 text-left text-sm font-light">Email</th>
                      <th className="px-4 py-3 text-left text-sm font-light">Role</th>
                      <th className="px-4 py-3 text-left text-sm font-light">Active</th>
                      <th className="px-4 py-3 text-left text-sm font-light">Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(users || []).map((u) => (
                      <tr key={u.user_id} className="border-b border-white/10 hover:bg-white/5">
                        <td className="px-4 py-3 text-sm font-light">{u.username}</td>
                        <td className="px-4 py-3 text-sm font-light">{u.email}</td>
                        <td className="px-4 py-3 text-sm font-light">
                          <select
                            value={u.role}
                            onChange={(e) => handleUpdateUserRole(u.user_id, e.target.value as "admin" | "user")}
                            className="bg-black border border-white rounded px-2 py-1 text-sm"
                          >
                            <option value="user">User</option>
                            <option value="admin">Admin</option>
                          </select>
                        </td>
                        <td className="px-4 py-3 text-sm font-light">{u.active ? "Yes" : "No"}</td>
                        <td className="px-4 py-3 text-sm font-light">
                          <button
                            onClick={() => setShowDeleteUserConfirm(u.user_id)}
                            className="text-red-400 hover:text-red-300 transition-colors"
                            title="Delete user"
                          >
                            <Trash2 className="h-4 w-4" />
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              {showDeleteUserConfirm && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
                  <div className="bg-black border border-white/10 rounded-lg p-6 max-w-md w-full mx-4">
                    <h3 className="text-lg font-light mb-4">Confirm Delete</h3>
                    <p className="text-white/60 font-light mb-6">
                      Are you sure you want to delete this user? This action cannot be undone.
                    </p>
                    <div className="flex gap-2">
                      <Button variant="destructive" onClick={() => handleDeleteUser(showDeleteUserConfirm)}>
                        Delete
                      </Button>
                      <Button variant="outline" onClick={() => setShowDeleteUserConfirm(null)}>
                        Cancel
                      </Button>
                    </div>
                  </div>
                </div>
              )}
            </div>
          )}

          {!loading && activeTab === "pis" && (
            <div>
              <div className="flex justify-between items-center mb-6">
                <h2 className="text-2xl font-light">PI Management</h2>
                <Button onClick={() => setShowPiForm(true)}>Create PI</Button>
              </div>

              {showPiForm && (
                <div className="border border-white/10 rounded-lg p-6 bg-black/50 mb-6">
                  <div className="flex justify-between items-center mb-4">
                    <h3 className="text-lg font-light">Create New PI</h3>
                    <button onClick={() => setShowPiForm(false)} className="text-white/60 hover:text-white">
                      <X className="h-5 w-5" />
                    </button>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <Input
                      placeholder="PI ID"
                      value={piFormData.pi_id}
                      onChange={(e) => setPiFormData({ ...piFormData, pi_id: e.target.value })}
                      className="bg-black/50 border-white/10"
                    />
                    <select
                      value={piFormData.user_id}
                      onChange={(e) => setPiFormData({ ...piFormData, user_id: e.target.value })}
                      className="flex h-9 w-full rounded-md border border-white bg-black px-3 py-1 text-sm"
                    >
                      <option value="">No User (Unassigned)</option>
                      {(users || []).map((u) => (
                        <option key={u.user_id} value={u.user_id}>
                          {u.username} ({u.email})
                        </option>
                      ))}
                    </select>
                  </div>
                  <div className="flex gap-2 mt-4">
                    <Button onClick={handleCreatePi}>Create</Button>
                    <Button variant="outline" onClick={() => setShowPiForm(false)}>Cancel</Button>
                  </div>
                </div>
              )}

              <div className="border border-white/10 rounded-lg overflow-hidden">
                <table className="w-full">
                  <thead className="bg-black/50 border-b border-white/10">
                    <tr>
                      <th className="px-4 py-3 text-left text-sm font-light">PI ID</th>
                      <th className="px-4 py-3 text-left text-sm font-light">User</th>
                      <th className="px-4 py-3 text-left text-sm font-light">View</th>
                      <th className="px-4 py-3 text-left text-sm font-light">Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(pis || []).map((pi) => (
                      <tr key={pi.pi_id} className="border-b border-white/10 hover:bg-white/5">
                        <td className="px-4 py-3 text-sm font-light">{pi.pi_id}</td>
                        <td className="px-4 py-3 text-sm font-light">
                          <select
                            value={pi.user_id || ""}
                            onChange={(e) => handleUpdatePi(pi.pi_id, e.target.value)}
                            className="bg-black border border-white rounded px-2 py-1 text-sm"
                          >
                            <option value="">Unassigned</option>
                            {(users || []).map((u) => (
                              <option key={u.user_id} value={u.user_id}>
                                {u.username}
                              </option>
                            ))}
                          </select>
                        </td>
                        <td className="px-4 py-3 text-sm font-light">
                          <Button
                            size="sm"
                            variant="outline"
                            className="text-black bg-white/90 border-2 border-white hover:bg-white/20 hover:border-white/30 text-xs font-light h-8 px-4  flex items-center gap-1.5"
                            onClick={() => {
                              setSelectedPi(pi);
                              setActiveTab("devices");
                              setShowDeleteDeviceConfirm(null);
                            }}
                          >
                            View
                          </Button>
                        </td>
                        <td className="px-4 py-3 text-sm font-light">
                          <button
                            onClick={() => setShowDeletePiConfirm(pi.pi_id)}
                            className="text-red-400 hover:text-red-300 transition-colors"
                            title="Delete PI"
                          >
                            <Trash2 className="h-4 w-4" />
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              {showDeletePiConfirm && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
                  <div className="bg-black border border-white/10 rounded-lg p-6 max-w-md w-full mx-4">
                    <h3 className="text-lg font-light mb-4">Confirm Delete</h3>
                    <p className="text-white/60 font-light mb-6">
                      Are you sure you want to delete this PI? This will also delete all associated devices and readings (cascade delete).
                    </p>
                    <div className="flex gap-2">
                      <Button variant="destructive" onClick={() => handleDeletePi(showDeletePiConfirm)}>
                        Delete
                      </Button>
                      <Button variant="outline" onClick={() => setShowDeletePiConfirm(null)}>
                        Cancel
                      </Button>
                    </div>
                  </div>
                </div>
              )}
            </div>
          )}

          {!loading && activeTab === "devices" && (
            <div>
              {!selectedPi ? (
                <>
                  <div className="flex justify-between items-center mb-6">
                    <div>
                      <h2 className="text-2xl font-light">All Devices</h2>
                      <p className="text-white/60 font-light text-sm mt-1">
                        {selectedPiFilter 
                          ? allDevices.filter(d => d.pi_id === selectedPiFilter).length 
                          : allDevices.length} device{(selectedPiFilter 
                          ? allDevices.filter(d => d.pi_id === selectedPiFilter).length 
                          : allDevices.length) !== 1 ? "s" : ""} {selectedPiFilter ? `for PI: ${selectedPiFilter}` : "across all PIs"}
                      </p>
                    </div>
                    <div className="flex gap-2 items-center">
                      <select
                        value={selectedPiFilter}
                        onChange={(e) => setSelectedPiFilter(e.target.value)}
                        className="flex h-9 rounded-md border border-white bg-black px-3 py-1 text-sm min-w-[200px]"
                      >
                        <option value="">All PIs</option>
                        {pis.map((pi) => (
                          <option key={pi.pi_id} value={pi.pi_id}>
                            {pi.pi_id}
                          </option>
                        ))}
                      </select>
                      <Button onClick={() => setActiveTab("pis")}>View PIs</Button>
                    </div>
                  </div>

                  {(() => {
                    const filteredDevices = selectedPiFilter 
                      ? allDevices.filter(d => d.pi_id === selectedPiFilter)
                      : allDevices;
                    
                    return filteredDevices.length === 0 ? (
                      <div className="border border-white/10 rounded-lg p-12 bg-black/50 text-center">
                        <p className="text-white/60 font-light mb-4">No devices found</p>
                        <p className="text-white/40 font-light text-sm">
                          {selectedPiFilter 
                            ? `No devices found for PI: ${selectedPiFilter}`
                            : "Devices will appear here once they are registered to a PI"}
                        </p>
                      </div>
                    ) : (
                      <div className="border border-white/10 rounded-lg overflow-hidden">
                        <table className="w-full">
                          <thead className="bg-black/50 border-b border-white/10">
                            <tr>
                              <th className="px-4 py-3 text-left text-sm font-light">PI ID</th>
                              <th className="px-4 py-3 text-left text-sm font-light">Device ID</th>
                              <th className="px-4 py-3 text-left text-sm font-light">View</th>
                              <th className="px-4 py-3 text-left text-sm font-light">Actions</th>
                            </tr>
                          </thead>
                          <tbody>
                            {filteredDevices.map((device) => (
                              <tr key={`${device.pi_id}-${device.device_id}`} className="border-b border-white/10 hover:bg-white/5">
                                <td className="px-4 py-3 text-sm font-light">
                                  <span className="text-white/80 font-mono">{device.pi_id}</span>
                                </td>
                                <td className="px-4 py-3 text-sm font-light font-mono">{device.device_id}</td>
                                <td className="px-4 py-3 text-sm font-light">
                                  <Button
                                    size="sm"
                                    variant="outline"
                                    className="text-xs font-light h-7 px-3"
                                    onClick={() => handleViewDeviceReadings(device.pi_id, device.device_id)}
                                  >
                                    View Readings
                                  </Button>
                                </td>
                                <td className="px-4 py-3 text-sm font-light">
                                  <button
                                    onClick={() =>
                                      setShowDeleteDeviceConfirm({ piId: device.pi_id, deviceId: device.device_id })
                                    }
                                    className="text-red-400 hover:text-red-300 transition-colors"
                                    title="Delete device"
                                  >
                                    <Trash2 className="h-4 w-4" />
                                  </button>
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    );
                  })()}
                </>
              ) : (
                <>
                  <div className="flex justify-between items-center mb-6">
                    <div>
                      <h2 className="text-2xl font-light">Device Management</h2>
                      <p className="text-white/60 font-light text-sm mt-1">PI: {selectedPi.pi_id}</p>
                    </div>
                    <div className="flex gap-2">
                      <Button variant="outline" onClick={() => {
                        setSelectedPi(null);
                        setShowDeleteDeviceConfirm(null);
                      }}>Back to PIs</Button>
                      <Button onClick={() => setShowDeviceForm(true)}>Create Device</Button>
                    </div>
                  </div>

                  {showDeviceForm && (
                    <div className="border border-white/10 rounded-lg p-6 bg-black/50 mb-6">
                      <div className="flex justify-between items-center mb-4">
                        <h3 className="text-lg font-light">Create New Device</h3>
                        <button onClick={() => setShowDeviceForm(false)} className="text-white/60 hover:text-white">
                          <X className="h-5 w-5" />
                        </button>
                      </div>
                      <div>
                        <label className="text-sm text-white/60 font-light mb-1 block">Device ID</label>
                        <Input
                          placeholder="Number (e.g., 1) or MAC Address (e.g., AA:BB:CC:DD:EE:FF)"
                          value={deviceFormData.device_id}
                          onChange={(e) => setDeviceFormData({ ...deviceFormData, device_id: e.target.value })}
                          className="bg-black/50 border-white/10"
                        />
                      </div>
                      <div className="flex gap-2 mt-4">
                        <Button 
                          onClick={handleCreateDevice}
                          disabled={!deviceFormData.device_id || !isValidDeviceId(deviceFormData.device_id)}
                        >
                          Create
                        </Button>
                        <Button variant="outline" onClick={() => setShowDeviceForm(false)}>Cancel</Button>
                      </div>
                    </div>
                  )}

                  <div className="border border-white/10 rounded-lg overflow-hidden">
                    <table className="w-full">
                      <thead className="bg-black/50 border-b border-white/10">
                        <tr>
                          <th className="px-4 py-3 text-left text-sm font-light">Device ID</th>
                          <th className="px-4 py-3 text-left text-sm font-light">View</th>
                          <th className="px-4 py-3 text-left text-sm font-light">Actions</th>
                        </tr>
                      </thead>
                      <tbody>
                        {(piDevices || []).map((device) => (
                          <tr key={`${device.pi_id}-${device.device_id}`} className="border-b border-white/10 hover:bg-white/5">
                            <td className="px-4 py-3 text-sm font-light">{device.device_id}</td>
                            <td className="px-4 py-3 text-sm font-light">
                              <Button
                                size="sm"
                                variant="outline"
                                className="text-xs font-light h-7 px-3"
                                onClick={() => handleViewDeviceReadings(device.pi_id, device.device_id)}
                              >
                                View Readings
                              </Button>
                            </td>
                            <td className="px-4 py-3 text-sm font-light">
                              <button
                                onClick={() =>
                                  setShowDeleteDeviceConfirm({ piId: device.pi_id, deviceId: device.device_id })
                                }
                                className="text-red-400 hover:text-red-300 transition-colors"
                                title="Delete device"
                              >
                                <Trash2 className="h-4 w-4" />
                              </button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>

                  {showDeleteDeviceConfirm && (
                    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
                      <div className="bg-black border border-white/10 rounded-lg p-6 max-w-md w-full mx-4">
                        <h3 className="text-lg font-light mb-4">Confirm Delete</h3>
                        <p className="text-white/60 font-light mb-6">
                          Are you sure you want to delete this device? This will also delete all associated readings (cascade delete).
                        </p>
                        <div className="flex gap-2">
                          <Button
                            variant="destructive"
                            onClick={() =>
                              handleDeleteDevice(showDeleteDeviceConfirm.piId, showDeleteDeviceConfirm.deviceId)
                            }
                          >
                            Delete
                          </Button>
                          <Button variant="outline" onClick={() => setShowDeleteDeviceConfirm(null)}>
                            Cancel
                          </Button>
                        </div>
                      </div>
                    </div>
                  )}
                </>
              )}
            </div>
          )}

          {!loading && activeTab === "readings" && (
            <div>
              {!showDeviceAnalytics ? (
                <>
              <div className="mb-6">
                <h2 className="text-2xl font-light mb-4">Reading Management</h2>
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
                    {(pis || []).map((pi) => (
                      <option key={pi.pi_id} value={pi.pi_id}>
                        {pi.pi_id}
                      </option>
                    ))}
                  </select>
                  {selectedPiForReadings && (
                    <select
                      value={selectedDeviceForReadings || ""}
                      onChange={(e) => {
                        const deviceId = e.target.value || null;
                        setSelectedDeviceForReadings(deviceId);
                            if (deviceId) {
                              handleDeviceClick(deviceId);
                            }
                      }}
                      className="flex h-9 w-full rounded-md border border-white bg-black px-3 py-1 text-sm"
                    >
                      <option value="">All Devices</option>
                      {(devices || [])
                        .filter((d) => d.pi_id === selectedPiForReadings)
                        .map((device) => (
                          <option key={device.device_id} value={String(device.device_id)}>
                            Device {device.device_id}
                          </option>
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
                        <th className="px-4 py-3 text-left text-sm font-light">Level</th>
                        <th className="px-4 py-3 text-left text-sm font-light">Battery</th>
                      </tr>
                    </thead>
                    <tbody>
                      {(readings || []).map((reading, idx) => (
                        <tr key={idx} className="border-b border-white/10 hover:bg-white/5">
                          <td className="px-4 py-3 text-sm font-light">
                            {new Date(reading.ts).toLocaleString()}
                          </td>
                              <td className="px-4 py-3 text-sm font-light">
                                <button
                                  onClick={() => handleDeviceClick(String(reading.device_id))}
                                  className="font-mono text-orange-400 hover:text-orange-300 hover:underline transition-colors cursor-pointer"
                                >
                                  {reading.device_id}
                                </button>
                              </td>
                          <td className="px-4 py-3 text-sm font-light">
                            {reading.payload.sensors.temperature
                              ? `${reading.payload.sensors.temperature.value} ${reading.payload.sensors.temperature.unit}`
                              : "N/A"}
                          </td>
                          <td className="px-4 py-3 text-sm font-light">
                            {reading.payload.sensors.level
                              ? `${reading.payload.sensors.level.value} ${reading.payload.sensors.level.unit}`
                              : "N/A"}
                          </td>
                          <td className="px-4 py-3 text-sm font-light">
                            {reading.payload.battery_percentage}%
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                  {readings.length === 0 && (
                    <div className="p-8 text-center text-white/60 font-light">
                      No readings found for the selected filters.
                    </div>
                  )}
                </div>
                  )}
                </>
              ) : (
                <>
                  {/* Device Analytics View */}
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
                        {(pis || []).map((pi) => (
                          <option key={pi.pi_id} value={pi.pi_id}>
                            {pi.pi_id}
                          </option>
                        ))}
                      </select>
                      {selectedPiForReadings && (
                        <select
                          value={selectedDeviceForReadings || ""}
                          onChange={(e) => {
                            const deviceId = e.target.value || null;
                            setSelectedDeviceForReadings(deviceId);
                            if (deviceId) {
                              handleDeviceClick(deviceId);
                            } else {
                              setShowDeviceAnalytics(false);
                            }
                          }}
                          className="flex h-9 w-full rounded-md border border-white bg-black px-3 py-1 text-sm"
                        >
                          <option value="">All Devices</option>
                          {(devices || [])
                            .filter((d) => d.pi_id === selectedPiForReadings)
                            .map((device) => (
                              <option key={device.device_id} value={String(device.device_id)}>
                                Device {device.device_id}
                              </option>
                            ))}
                        </select>
                      )}
                    </div>
                  </div>

                  {isLoadingReadings ? (
                    <div className="flex items-center justify-center py-12">
                      <Loader2 className="h-6 w-6 text-white/60 animate-spin" />
                    </div>
                  ) : (
                    <>
                      {/* Latest Reading Card */}
                      {latestReading && (
                        <div className="mb-6 border border-white/10 rounded-lg p-6 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm">
                          <div className="flex items-center justify-between mb-4">
                            <h2 className="text-xl font-light">Current Reading</h2>
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={refreshDeviceAnalytics}
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
                                  {new Date(latestReading.ts).toLocaleString("en-US", {
                                    month: "short",
                                    day: "numeric",
                                    hour: "2-digit",
                                    minute: "2-digit",
                                  })}
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
                                  {new Date(latestReading.ts).toLocaleString("en-US", {
                                    month: "short",
                                    day: "numeric",
                                    hour: "2-digit",
                                    minute: "2-digit",
                                  })}
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
                                {new Date(latestReading.ts).toLocaleString("en-US", {
                                  month: "short",
                                  day: "numeric",
                                  hour: "2-digit",
                                  minute: "2-digit",
                                })}
                              </div>
                            </div>
                          </div>
                        </div>
                      )}

                      {/* Readings Chart */}
                      {readingsForChart && Array.isArray(readingsForChart) && readingsForChart.length > 0 && (
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
                            <ReadingsChart readings={readingsForChart} timeRange={timeRange} />
                          </div>
                        </div>
                      )}

                      {/* Readings Table */}
                      {readingsForChart && Array.isArray(readingsForChart) && readingsForChart.length > 0 && (
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
                                <th className="px-4 py-3 text-left text-sm font-light">Level</th>
                                <th className="px-4 py-3 text-left text-sm font-light">Battery</th>
                              </tr>
                            </thead>
                            <tbody>
                              {[...readingsForChart]
                                .sort((a, b) => new Date(b.ts).getTime() - new Date(a.ts).getTime())
                                .map((reading, idx) => (
                                <tr key={idx} className="border-b border-white/10 hover:bg-white/5">
                                  <td className="px-4 py-3 text-sm font-light">
                                    {new Date(reading.ts).toLocaleString()}
                                  </td>
                                  <td className="px-4 py-3 text-sm font-light font-mono">{reading.device_id}</td>
                                  <td className="px-4 py-3 text-sm font-light">
                                    {reading.payload.sensors.temperature
                                      ? `${reading.payload.sensors.temperature.value} ${reading.payload.sensors.temperature.unit}`
                                      : "N/A"}
                                  </td>
                                  <td className="px-4 py-3 text-sm font-light">
                                    {reading.payload.sensors.level
                                      ? `${reading.payload.sensors.level.value} ${reading.payload.sensors.level.unit}`
                                      : "N/A"}
                                  </td>
                                  <td className="px-4 py-3 text-sm font-light">
                                    {reading.payload.battery_percentage}%
                                  </td>
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
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
