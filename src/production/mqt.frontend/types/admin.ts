export type UserRole = "user" | "admin";

export interface User {
  user_id: string;
  username: string;
  email: string;
  role: UserRole;
  active: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface Pi {
  pi_id: string;
  user_id?: string;
  created_at: string;
  updated_at?: string;
}

export interface Device {
  device_id: number | string; // Can be numeric ID or MAC address string (e.g., "AA:BB:CC:DD:EE:FF")
  pi_id: string;
  created_at: string;
  updated_at?: string;
}

export interface SensorReading {
  value: number;
  unit: string;
}

export interface SensorData {
  temperature?: SensorReading;
  humidity?: SensorReading;
  level?: SensorReading;
  light?: SensorReading;
  pressure?: SensorReading;
}

export interface Reading {
  ts: string;
  device_id: number | string; // Can be numeric ID or MAC address string
  pi_id: string;
  payload: {
    sensors: SensorData;
    battery_percentage: number;
  };
}

export interface SummaryStatistics {
  total_readings: number;
  avg_temperature?: number;
  avg_humidity?: number;
  min_temperature?: number;
  max_temperature?: number;
}

export interface PaginatedResponse<T> {
  items: T[];
  total?: number;
  page?: number;
  page_size?: number;
  next_page?: number | null;
}

export interface CreatePiRequest {
  pi_id: string;
  user_id?: string;
}

export interface UpdatePiRequest {
  user_id?: string;
}

export interface CreateDeviceRequest {
  device_id: string | number; // Accepts MAC address as string (e.g., "AA:BB:CC:DD:EE:FF") or numeric ID
}

export interface UpdateDeviceRequest {
  // No fields - device updates not currently supported
}

export interface UpdateUserRequest {
  username?: string;
  email?: string;
  active?: boolean;
}

export interface UpdateUserRoleRequest {
  role: UserRole;
}

export interface RegisterAdminRequest {
  username: string;
  email: string;
  password: string;
  role: UserRole;
}

export interface GetReadingsParams {
  pi_id?: string;
  device_id?: number;
  page?: number;
  page_size?: number;
  start_date?: string;
  end_date?: string;
}

export interface GetDeviceReadingsParams {
  page?: number;
  page_size?: number;
  start_date?: string;
  end_date?: string;
}

export interface GetStatsParams {
  pi_id?: string;
  device_id?: number;
  start_date?: string;
  end_date?: string;
}

