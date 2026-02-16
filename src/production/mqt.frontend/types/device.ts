export interface SensorReading {
  value: number;
  unit: string;
}

export interface SensorData {
  temperature?: SensorReading;
  level?: SensorReading;
}

export interface MQTTPayload {
  device_id: string;
  pi_id: string;
  timestamp: string;
  sensors: SensorData;
  battery_percentage: number;
}

export interface MQTTEnvelope {
  topic: string;
  payload: MQTTPayload;
  qos: number;
  retain: boolean;
  message_id: number;
  duplicate: boolean;
}

export interface Device {
  id: string;
  device_id: string;
  pi_id: string;
  name?: string;
  latitude: number;
  longitude: number;
  color?: string; // Hex color code from backend
  lastReading?: MQTTPayload;
  createdAt: string;
  updatedAt: string;
}

export interface CreateDeviceRequest {
  device_id: string;
  pi_id: string;
  latitude: number;
  longitude: number;
}

