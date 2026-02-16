"use client";

import { useState, useEffect } from "react";
import { useRequireAuth } from "@/hooks/useRequireAuth";
import Navbar from "@/components/navbar/Navbar";
import { LocationInput } from "@/components/weather/LocationInput";
import { CurrentWeather } from "@/components/weather/CurrentWeather";
import { WeatherForecast } from "@/components/weather/WeatherForecast";
import { weatherService } from "@/services/api/weatherService";
import type { CurrentWeather as CurrentWeatherType, WeatherForecast as WeatherForecastType } from "@/types/weather";
import { Loader2, AlertCircle, MapPin } from "lucide-react";

export default function ForecastPage() {
  const { user, isLoading } = useRequireAuth("user");
  const [hasLocation, setHasLocation] = useState<boolean | null>(null);
  const [isCheckingLocation, setIsCheckingLocation] = useState(true);
  const [isUpdatingLocation, setIsUpdatingLocation] = useState(false);
  const [currentWeather, setCurrentWeather] = useState<CurrentWeatherType | null>(null);
  const [forecast, setForecast] = useState<WeatherForecastType | null>(null);
  const [isLoadingWeather, setIsLoadingWeather] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [locationName, setLocationName] = useState<string>("");

  // Check if user has location set
  useEffect(() => {
    if (!isLoading && user) {
      checkLocation();
    }
  }, [isLoading, user]);

  // Fetch weather data when location is available
  useEffect(() => {
    if (hasLocation === true) {
      fetchWeatherData();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [hasLocation]);

  const checkLocation = async () => {
    try {
      setIsCheckingLocation(true);
      const profile = await weatherService.getProfile();
      if (profile && profile.latitude !== null && profile.latitude !== undefined && 
          profile.longitude !== null && profile.longitude !== undefined) {
        setHasLocation(true);
      } else {
        setHasLocation(false);
      }
    } catch (err) {
      console.error("Error checking location:", err);
      setError("Failed to check location status");
      setHasLocation(false);
    } finally {
      setIsCheckingLocation(false);
    }
  };

  const handleLocationSubmit = async (lat: number, lon: number, name: string) => {
    try {
      setIsUpdatingLocation(true);
      setError(null);
      await weatherService.updateLocation({ latitude: lat, longitude: lon });
      setLocationName(name);
      setHasLocation(true);
    } catch (err) {
      console.error("Error updating location:", err);
      const errorMessage = err instanceof Error ? err.message : "Failed to update location";
      setError(errorMessage);
    } finally {
      setIsUpdatingLocation(false);
    }
  };

  const fetchWeatherData = async () => {
    try {
      setIsLoadingWeather(true);
      setError(null);
      const [current, forecastData] = await Promise.all([
        weatherService.getCurrentWeather(),
        weatherService.getForecast(),
      ]);
      setCurrentWeather(current);
      setForecast(forecastData);
      if (!locationName && current.name) {
        setLocationName(current.name);
      }
    } catch (err) {
      console.error("Error fetching weather:", err);
      const errorMessage = err instanceof Error ? err.message : "Failed to fetch weather data";
      if (errorMessage.includes("location not set")) {
        setHasLocation(false);
        setError("Location not set. Please set your location first.");
      } else {
        setError(errorMessage);
      }
    } finally {
      setIsLoadingWeather(false);
    }
  };

  if (isLoading || isCheckingLocation) {
    return (
      <div className="min-h-screen bg-black text-white flex items-center justify-center">
        <div className="flex flex-col items-center gap-3">
          <Loader2 className="h-6 w-6 text-white/60 animate-spin" />
          <p className="text-white/60 font-light">Loading...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-black text-white">
      <Navbar />
      <main className="pt-24 px-4 sm:px-6 lg:px-8 pb-16">
        <div className="max-w-7xl mx-auto">
          <div className="mb-8">
            <h1 className="text-4xl font-light tracking-tight mb-2">
              Weather Forecast
            </h1>
            {/* <p className="text-white/60 font-light text-sm">
              View current conditions and extended forecast
            </p> */}
          </div>

          {error && (
            <div className="mb-6 p-4 border border-red-500/20 bg-red-500/10 rounded-lg flex items-center gap-2">
              <AlertCircle className="h-5 w-5 text-red-400" />
              <p className="text-sm text-red-400 font-light">{error}</p>
            </div>
          )}

          {/* Location Input Panel */}
          {/* <div className="mb-6 border border-white/10 rounded-lg p-6 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm">
            <h2 className="text-xl font-light mb-4">Location Settings</h2>
            {locationName && (
              <p className="text-white/60 font-light text-sm mb-4">
                Current location: <span className="text-white">{locationName}</span>
              </p>
            )}
            <LocationInput
              onLocationSubmit={handleLocationSubmit}
              isLoading={isUpdatingLocation}
            />
          </div> */}

          {!hasLocation ? (
            <div className="border border-white/10 rounded-lg p-12 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm text-center">
              <MapPin className="h-12 w-12 text-white/40 mx-auto mb-4" />
              <p className="text-white/60 font-light">
                Please set your location to view weather information. This can be done in the <a href="/user/settings" className="text-orange-400 hover:text-orange-300">settings</a> page.
              </p>
            </div>
          ) : (
            <div className="space-y-6">
              {isLoadingWeather ? (
                <div className="flex items-center justify-center py-12">
                  <div className="flex flex-col items-center gap-3">
                    <Loader2 className="h-8 w-8 text-white/60 animate-spin" />
                    <p className="text-white/60 font-light">Loading weather data...</p>
                  </div>
                </div>
              ) : (
                <>
                  {currentWeather && <CurrentWeather weather={currentWeather} />}
                  {forecast && <WeatherForecast forecast={forecast} />}
                </>
              )}
            </div>
          )}
        </div>
      </main>
    </div>
  );
}

