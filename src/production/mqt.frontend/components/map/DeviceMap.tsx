"use client";

import { useState, useEffect, useCallback, useRef, useMemo } from "react";
import Map, { Marker, Popup } from "react-map-gl/maplibre";
import type { MapRef } from "react-map-gl/maplibre";
import "maplibre-gl/dist/maplibre-gl.css";
import { Plus, X, MapPin, Navigation, Crosshair, Search, User } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { Device } from "@/types/device";
import type { Pi, PiDevice } from "@/services/api/sensorService";
import { usePiPreferences, colorToGradient } from "@/hooks/usePiPreferences";
import { MarkerShapeComponent } from "./MarkerShape";

interface DeviceMapProps {
  devices: Device[];
  onDeviceAdd: (device: Omit<Device, "id" | "createdAt" | "updatedAt">) => void;
  onDeviceClick: (device: Device) => void;
  center?: [number, number];
  availablePis?: Pi[];
  availableDevices?: PiDevice[];
  selectedDeviceId?: string | null;
  onCenterDevice?: (device: Device) => void;
  carousel?: React.ReactNode;
}

export function DeviceMap({ devices, onDeviceAdd, onDeviceClick, center = [40.7128, -74.006], availablePis = [], availableDevices = [], selectedDeviceId = null, onCenterDevice, carousel }: DeviceMapProps) {
  const { getPreference, initializePreferences, loadColorsFromBackend } = usePiPreferences();
  const [isAddingDevice, setIsAddingDevice] = useState(false);
  const [showAddForm, setShowAddForm] = useState(false);
  const [newDevice, setNewDevice] = useState({
    device_id: "",
    pi_id: "",
    latitude: "",
    longitude: "",
  });
  
  // Filter devices based on selected PI
  const filteredDevices = useMemo(() => {
    if (!newDevice.pi_id) return [];
    return availableDevices.filter(
      (device) => device.pi_id === newDevice.pi_id
    );
  }, [availableDevices, newDevice.pi_id]);

  // Create a stable key from PI IDs for dependency tracking
  const piIdsKey = useMemo(() => {
    return availablePis.map(pi => pi.pi_id).sort().join(',');
  }, [availablePis]);

  // Initialize preferences for available PIs when they're loaded
  useEffect(() => {
    if (availablePis.length > 0) {
      initializePreferences(availablePis.map(pi => pi.pi_id));
      // Load colors from backend device locations
      loadColorsFromBackend();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [piIdsKey]); // Only re-run when PI IDs actually change
  const [tempMarker, setTempMarker] = useState<[number, number] | null>(null);
  const [mounted, setMounted] = useState(false);
  const [viewState, setViewState] = useState({
    longitude: center[1],
    latitude: center[0],
    zoom: 13,
  });
  const [clickedLocation, setClickedLocation] = useState<{ lat: number; lng: number } | null>(null);
  const [showCenterInput, setShowCenterInput] = useState(false);
  const [centerInput, setCenterInput] = useState({ lat: "", lng: "" });
  const [mapError, setMapError] = useState<string | null>(null);
  const [formError, setFormError] = useState<string | null>(null);
  const [currentLocation, setCurrentLocation] = useState<{ lat: number; lng: number } | null>(null);
  const mapRef = useRef<MapRef>(null);

  useEffect(() => {
    try {
      setMounted(true);
      setMapError(null);
      setViewState({
        longitude: center[1],
        latitude: center[0],
        zoom: 13,
      });
    } catch (error) {
      console.error("Error initializing map:", error);
      setMapError(error instanceof Error ? error.message : "Failed to initialize map");
    }
  }, [center]);

  const centerMap = useCallback((lat: number, lng: number, zoom: number = 13) => {
    const newViewState = {
      longitude: lng,
      latitude: lat,
      zoom,
    };
    setViewState(newViewState);
    if (mapRef.current) {
      mapRef.current.flyTo({
        center: [lng, lat],
        zoom,
        duration: 1000,
      });
    }
  }, []);

  const handleSetToCurrentLocation = useCallback(() => {
    // Check if geolocation is supported
    if (!navigator.geolocation) {
      alert("Geolocation is not supported by your browser. Please use the search feature to set your location.");
      return;
    }

    // Check if we're in a secure context (HTTPS or localhost)
    if (!window.isSecureContext) {
      alert("Geolocation requires a secure connection (HTTPS). Please use the search feature to set your location.");
      return;
    }

    navigator.geolocation.getCurrentPosition(
      (position) => {
        const { latitude, longitude } = position.coords;
        setCurrentLocation({ lat: latitude, lng: longitude });
        centerMap(latitude, longitude);
      },
      (error: GeolocationPositionError) => {
        console.error("Error getting current location:", {
          code: error.code,
          message: error.message,
          error
        });
        
        let errorMessage = "Unable to get your current location.";
        
        // Handle specific error codes
        switch (error.code) {
          case error.PERMISSION_DENIED:
            errorMessage = "Location access denied. Please enable location permissions in your browser settings and try again.";
            break;
          case error.POSITION_UNAVAILABLE:
            errorMessage = "Location information is unavailable. Your device may not be able to determine its location. Please try again or use the search feature.";
            break;
          case error.TIMEOUT:
            errorMessage = "Location request timed out. Please try again or use the search feature to set your location.";
            break;
          default:
            errorMessage = `Unable to get your current location: ${error.message || "Unknown error"}. Please check your browser settings or use the search feature.`;
        }
        
        alert(errorMessage);
      },
      {
        enableHighAccuracy: true,
        timeout: 10000, // 10 seconds
        maximumAge: 60000 // Accept cached position up to 1 minute old
      }
    );
  }, [centerMap]);

  const handleSetToCenter = useCallback(() => {
    const lat = parseFloat(centerInput.lat);
    const lng = parseFloat(centerInput.lng);
    
    if (isNaN(lat) || isNaN(lng)) {
      alert("Please enter valid latitude and longitude values.");
      return;
    }
    
    if (lat < -90 || lat > 90) {
      alert("Latitude must be between -90 and 90.");
      return;
    }
    
    if (lng < -180 || lng > 180) {
      alert("Longitude must be between -180 and 180.");
      return;
    }
    
    centerMap(lat, lng);
    setShowCenterInput(false);
    setCenterInput({ lat: "", lng: "" });
  }, [centerInput, centerMap]);

  const handleResetToDefault = useCallback(() => {
    centerMap(center[0], center[1]);
  }, [center, centerMap]);

  // Center map on selected device
  useEffect(() => {
    if (selectedDeviceId && mapRef.current) {
      const device = devices.find((d) => d.id === selectedDeviceId);
      if (device) {
        centerMap(device.latitude, device.longitude, 15);
      }
    }
  }, [selectedDeviceId, devices, centerMap]);

  const handleMapClick = useCallback((event: { lngLat: { lat: number; lng: number } }) => {
    if (isAddingDevice) {
      const { lat, lng } = event.lngLat;
      setTempMarker([lat, lng]);
      setClickedLocation({ lat, lng });
      setNewDevice((prev) => ({
        ...prev,
        latitude: lat.toFixed(6),
        longitude: lng.toFixed(6),
      }));
      setShowAddForm(true);
    }
  }, [isAddingDevice]);

  const handleAddDevice = () => {
    if (!newDevice.device_id || !newDevice.pi_id || !newDevice.latitude || !newDevice.longitude) {
      return;
    }

    // Check if device already exists on the map
    const deviceExists = devices.some(
      (device) => device.device_id === newDevice.device_id && device.pi_id === newDevice.pi_id
    );

    if (deviceExists) {
      setFormError("This device is already marked on the map. Please select a different device.");
      return;
    }

    // Clear any previous errors
    setFormError(null);

    onDeviceAdd({
      device_id: newDevice.device_id,
      pi_id: newDevice.pi_id,
      latitude: parseFloat(newDevice.latitude),
      longitude: parseFloat(newDevice.longitude),
    });

    // Reset form
    setNewDevice({
      device_id: "",
      pi_id: "",
      latitude: "",
      longitude: "",
    });
    setShowAddForm(false);
    setIsAddingDevice(false);
    setTempMarker(null);
    setClickedLocation(null);
    setFormError(null);
  };

  const handleCancel = () => {
    setShowAddForm(false);
    setIsAddingDevice(false);
    setTempMarker(null);
    setClickedLocation(null);
    setFormError(null);
    setNewDevice({
      device_id: "",
      pi_id: "",
      latitude: "",
      longitude: "",
    });
  };

  const handlePiChange = (piId: string) => {
    setFormError(null); // Clear error when PI changes
    setNewDevice({
      ...newDevice,
      pi_id: piId,
      device_id: "", // Reset device when PI changes
    });
  };

  if (!mounted) {
    return (
      <div className="relative w-full h-[600px] rounded-lg overflow-hidden border border-white/10 flex items-center justify-center bg-black/50">
        <div className="text-white/60 font-light">Loading map...</div>
      </div>
    );
  }

  if (mapError) {
    return (
      <div className="relative w-full h-[600px] rounded-lg overflow-hidden border border-red-500/20 bg-red-500/10 flex items-center justify-center">
        <div className="text-center">
          <div className="text-red-400 font-light mb-2">Map Error</div>
          <div className="text-white/60 text-sm font-light mb-4">{mapError}</div>
          <button
            onClick={() => {
              setMapError(null);
              setMounted(false);
              setTimeout(() => setMounted(true), 100);
            }}
            className="px-4 py-2 bg-white/10 hover:bg-white/20 border border-white/20 text-white rounded-lg text-sm font-light transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="relative w-full h-[600px] rounded-lg overflow-hidden border border-white/10">
      <Map
        ref={mapRef}
        {...viewState}
        onMove={(evt) => setViewState(evt.viewState)}
        onClick={handleMapClick}
        mapStyle="https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json"
        style={{ width: "100%", height: "100%" }}
        cursor={isAddingDevice ? "crosshair" : "default"}
      >
        {/* Existing device markers */}
        {devices.map((device, index) => {
          const isSelected = device.id === selectedDeviceId;
          const preference = getPreference(device.pi_id, index);
          const markerSize = isSelected ? "lg" : "md";
          
          return (
            <Marker
              key={device.id}
              longitude={device.longitude}
              latitude={device.latitude}
              anchor="bottom"
              onClick={(e) => {
                e.originalEvent.stopPropagation();
                onDeviceClick(device);
              }}
            >
              <div className="cursor-pointer">
                <MarkerShapeComponent
                  gradient={colorToGradient(preference.color)}
                  size={markerSize}
                  isSelected={isSelected}
                />
              </div>
            </Marker>
          );
        })}

        {/* Current location marker (blue human icon) */}
        {currentLocation && (
          <Marker
            longitude={currentLocation.lng}
            latitude={currentLocation.lat}
            anchor="bottom"
          >
            <div className="w-8 h-8 bg-gradient-to-br from-blue-400 to-blue-500 rounded-full border-2 border-blue-300 shadow-lg shadow-blue-400/70 flex items-center justify-center">
              <User className="w-5 h-5 text-white" strokeWidth={2} />
            </div>
          </Marker>
        )}

        {/* Temporary marker for new device placement */}
        {tempMarker && (
          <Marker
            longitude={tempMarker[1]}
            latitude={tempMarker[0]}
            anchor="bottom"
          >
            <div className="w-6 h-6 bg-gradient-to-br from-orange-400 to-orange-500 rounded-full border-2 border-white shadow-lg shadow-orange-400/50 flex items-center justify-center animate-pulse">
              <div className="w-2 h-2 bg-white rounded-full" />
            </div>
          </Marker>
        )}

        {/* Popup for clicked location */}
        {clickedLocation && isAddingDevice && (
          <Popup
            longitude={clickedLocation.lng}
            latitude={clickedLocation.lat}
            anchor="bottom"
            onClose={() => setClickedLocation(null)}
            closeButton={false}
          >
            <div className="text-black text-xs">
              <p>Click to place device here</p>
            </div>
          </Popup>
        )}
      </Map>

      {/* Map Centering Controls */}
      <div className="absolute top-4 left-4 z-[1000] flex flex-col gap-2">
        <div className="flex flex-col gap-2 bg-black/80 backdrop-blur-sm border border-white/10 rounded-lg p-2">
          <Button
            onClick={handleSetToCurrentLocation}
            className="bg-white/10 hover:bg-orange-500/20 border border-white/20 hover:border-orange-500/40 text-white hover:text-orange-200 backdrop-blur-sm transition-all"
            size="sm"
            title="Center map to your current location"
          >
            <Navigation className="h-4 w-4" />
          </Button>
          <Button
            onClick={() => setShowCenterInput(!showCenterInput)}
            className="bg-white/10 hover:bg-orange-500/20 border border-white/20 hover:border-orange-500/40 text-white hover:text-orange-200 backdrop-blur-sm transition-all"
            size="sm"
            title="Set map to specific coordinates"
          >
            <Search className="h-4 w-4" />
          </Button>
          <Button
            onClick={handleResetToDefault}
            className="bg-white/10 hover:bg-orange-500/20 border border-white/20 hover:border-orange-500/40 text-white hover:text-orange-200 backdrop-blur-sm transition-all"
            size="sm"
            title="Reset to default center"
          >
            <Crosshair className="h-4 w-4" />
          </Button>
        </div>
        {showCenterInput && (
          <div className="bg-black/95 border border-white/10 rounded-lg p-4 backdrop-blur-md min-w-[280px]">
            <h4 className="text-sm font-light mb-3 text-white">Set Map Center</h4>
            <div className="space-y-2">
              <div>
                <label className="text-xs text-white/60 font-light mb-1 block">Latitude</label>
                <Input
                  type="number"
                  step="any"
                  value={centerInput.lat}
                  onChange={(e) => setCenterInput({ ...centerInput, lat: e.target.value })}
                  placeholder="40.7128"
                  className="bg-white/5 border-white/10 text-white"
                />
              </div>
              <div>
                <label className="text-xs text-white/60 font-light mb-1 block">Longitude</label>
                <Input
                  type="number"
                  step="any"
                  value={centerInput.lng}
                  onChange={(e) => setCenterInput({ ...centerInput, lng: e.target.value })}
                  placeholder="-74.0060"
                  className="bg-white/5 border-white/10 text-white"
                />
              </div>
              <div className="flex gap-2 pt-2">
                <Button
                  onClick={handleSetToCenter}
                  className="flex-1 bg-white/10 hover:bg-orange-500/20 border border-white/20 hover:border-orange-500/40 text-white hover:text-orange-200 transition-all"
                  size="sm"
                >
                  Center
                </Button>
                <Button
                  onClick={() => {
                    setShowCenterInput(false);
                    setCenterInput({ lat: "", lng: "" });
                  }}
                  variant="outline"
                  className="border-white/20 text-white hover:bg-white/10"
                  size="sm"
                >
                  Cancel
                </Button>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Add Device Button */}
      <div className="absolute top-4 right-4 z-[1000]">
        {!isAddingDevice ? (
          <Button
            onClick={() => setIsAddingDevice(true)}
            className="bg-white/10 hover:bg-orange-500/20 border border-white/20 hover:border-orange-500/40 text-white hover:text-orange-200 backdrop-blur-sm transition-all"
          >
            <Plus className="h-4 w-4 mr-2" />
            Add Device
          </Button>
        ) : (
          <Button
            onClick={handleCancel}
            className="bg-red-500/20 hover:bg-red-500/30 border border-red-500/30 text-white backdrop-blur-sm"
          >
            <X className="h-4 w-4 mr-2" />
            Cancel
          </Button>
        )}
      </div>

      {/* Add Device Form */}
      {showAddForm && (
        <div className="absolute bottom-4 left-4 right-4 z-[1000] bg-black/95 border border-white/10 rounded-lg p-6 backdrop-blur-md">
          <h3 className="text-lg font-light mb-4">Add New Device</h3>
          <div className="space-y-3">
            <div>
              <label className="text-xs text-white/60 font-light mb-1 block">Raspberry Pi Unit</label>
              {availablePis.length > 0 ? (
                <select
                  value={newDevice.pi_id}
                  onChange={(e) => handlePiChange(e.target.value)}
                  className="flex h-10 w-full rounded-md border border-white/10 bg-white/5 px-3 py-2 text-sm text-white placeholder:text-white/40 focus:outline-none focus:ring-2 focus:ring-orange-500/50 focus:border-orange-500/50 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  <option value="">Select a Raspberry Pi Unit</option>
                  {availablePis.map((pi) => (
                    <option key={pi.pi_id} value={pi.pi_id}>
                      {pi.pi_id}
                    </option>
                  ))}
                </select>
              ) : (
                <Input
                  value={newDevice.pi_id}
                  onChange={(e) => setNewDevice({ ...newDevice, pi_id: e.target.value })}
                  placeholder="e.g., Pi-001"
                  className="bg-white/5 border-white/10 text-white"
                />
              )}
            </div>
            <div>
              <label className="text-xs text-white/60 font-light mb-1 block">Device ID (MAC Address)</label>
              {newDevice.pi_id && filteredDevices.length > 0 ? (
                <select
                  value={newDevice.device_id}
                  onChange={(e) => {
                    setFormError(null); // Clear error when device changes
                    setNewDevice({ ...newDevice, device_id: e.target.value });
                  }}
                  className="flex h-10 w-full rounded-md border border-white/10 bg-white/5 px-3 py-2 text-sm text-white placeholder:text-white/40 focus:outline-none focus:ring-2 focus:ring-orange-500/50 focus:border-orange-500/50 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  <option value="">Select a Device</option>
                  {filteredDevices.map((device) => (
                    <option key={device.device_id} value={device.device_id}>
                      {device.device_id}
                    </option>
                  ))}
                </select>
              ) : (
                <Input
                  value={newDevice.device_id}
                  onChange={(e) => {
                    setFormError(null); // Clear error when device changes
                    setNewDevice({ ...newDevice, device_id: e.target.value });
                  }}
                  placeholder={newDevice.pi_id ? "No devices available for this PI" : "Select a PI first"}
                  className="bg-white/5 border-white/10 text-white"
                  disabled={!newDevice.pi_id}
                />
              )}
            </div>
            {formError && (
              <div className="p-3 border border-red-500/20 bg-red-500/10 rounded-md">
                <p className="text-sm text-red-400 font-light">{formError}</p>
              </div>
            )}
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-xs text-white/60 font-light mb-1 block">Latitude</label>
                <Input
                  type="number"
                  step="any"
                  value={newDevice.latitude}
                  onChange={(e) => setNewDevice({ ...newDevice, latitude: e.target.value })}
                  placeholder="40.7128"
                  className="bg-white/5 border-white/10 text-white"
                />
              </div>
              <div>
                <label className="text-xs text-white/60 font-light mb-1 block">Longitude</label>
                <Input
                  type="number"
                  step="any"
                  value={newDevice.longitude}
                  onChange={(e) => setNewDevice({ ...newDevice, longitude: e.target.value })}
                  placeholder="-74.0060"
                  className="bg-white/5 border-white/10 text-white"
                />
              </div>
            </div>
            <div className="flex gap-2 pt-2">
              <Button
                onClick={handleAddDevice}
                className="flex-1 bg-white/10 hover:bg-orange-500/20 border border-white/20 hover:border-orange-500/40 text-white hover:text-orange-200 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                disabled={!newDevice.device_id || !newDevice.pi_id || !newDevice.latitude || !newDevice.longitude}
              >
                <MapPin className="h-4 w-4 mr-2" />
                Add Device
              </Button>
              <Button
                onClick={handleCancel}
                variant="outline"
                className="border-white/20 text-white hover:bg-white/10"
              >
                Cancel
              </Button>
            </div>
            <p className="text-xs text-white/40 font-light">
              {isAddingDevice ? "Click on the map to place device, or enter coordinates manually" : ""}
            </p>
          </div>
        </div>
      )}

      {/* Carousel inside map */}
      {carousel && (
        <div className="absolute bottom-4 left-4 z-[1000]">
          {carousel}
        </div>
      )}
    </div>
  );
}
