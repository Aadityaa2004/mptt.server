"use client";

import { useMemo } from "react";
import Map, { Marker, NavigationControl } from "react-map-gl/maplibre";
import "maplibre-gl/dist/maplibre-gl.css";
import type { UserLocation } from "@/types/admin";

const MAP_STYLE = "https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json";

interface UserDetailMapProps {
  locations: UserLocation[];
}

export function UserDetailMap({ locations }: UserDetailMapProps) {
  const center = useMemo(() => {
    if (locations.length === 0) return { lng: -74.006, lat: 40.7128 };
    const lng = locations.reduce((s, l) => s + l.longitude, 0) / locations.length;
    const lat = locations.reduce((s, l) => s + l.latitude, 0) / locations.length;
    return { lng, lat };
  }, [locations]);

  if (locations.length === 0) {
    return (
      <div className="w-full h-full flex items-center justify-center bg-white/5 text-white/40 font-light">
        No locations to display
      </div>
    );
  }

  return (
    <Map
      mapStyle={MAP_STYLE}
      initialViewState={{
        longitude: center.lng,
        latitude: center.lat,
        zoom: 10,
      }}
      style={{ width: "100%", height: "100%" }}
    >
      <NavigationControl position="top-right" />
      {locations.map((loc, i) => (
        <Marker
          key={`${loc.pi_id}-${loc.device_id}-${i}`}
          longitude={loc.longitude}
          latitude={loc.latitude}
          anchor="bottom"
        >
          <div
            className="w-6 h-6 rounded-full border-2 border-white shadow-lg"
            style={{ backgroundColor: loc.color || "#ea580c" }}
            title={`Device ${loc.device_id} (${loc.pi_id})`}
          />
        </Marker>
      ))}
    </Map>
  );
}
