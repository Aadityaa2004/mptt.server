"use client";

import { CurrentWeather as CurrentWeatherType } from "@/types/weather";
import { Droplets, Wind, Eye, Gauge } from "lucide-react";

interface CurrentWeatherProps {
  weather: CurrentWeatherType;
}

// Convert Celsius to Fahrenheit
const celsiusToFahrenheit = (celsius: number): number => {
  return (celsius * 9) / 5 + 32;
};

export function CurrentWeather({ weather }: CurrentWeatherProps) {
  const mainCondition = weather.weather[0];
  const iconUrl = `https://openweathermap.org/img/wn/${mainCondition.icon}@2x.png`;

  const formatTime = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleTimeString("en-US", {
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  return (
    <div className="select-none">
      <div className="flex flex-wrap items-center gap-x-8 gap-y-2">
        <div className="flex items-center gap-4">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src={iconUrl} alt={mainCondition.main} className="w-11 h-11" />
          <div>
            <div className="text-2xl font-light">{Math.round(celsiusToFahrenheit(weather.main.temp))}°F</div>
            <p className="text-white/50 text-sm font-light capitalize">{weather.name} · {mainCondition.description}</p>
          </div>
        </div>
        <div className="h-6 w-px bg-white/10 hidden sm:block" />
        <div className="flex items-center gap-6 text-sm">
          <div className="flex items-center gap-2">
            <Droplets className="h-4 w-4 text-white/40" />
            <span className="text-white/70 font-light">{weather.main.humidity}%</span>
            <span className="text-white/40 font-light text-xs">humidity</span>
          </div>
          <div className="flex items-center gap-2">
            <Wind className="h-4 w-4 text-white/40" />
            <span className="text-white/70 font-light">{weather.wind.speed} m/s</span>
            <span className="text-white/40 font-light text-xs">wind</span>
          </div>
          <div className="flex items-center gap-2">
            <Gauge className="h-4 w-4 text-white/40" />
            <span className="text-white/70 font-light">{weather.main.pressure} hPa</span>
          </div>
          <div className="flex items-center gap-2">
            <Eye className="h-4 w-4 text-white/40" />
            <span className="text-white/70 font-light">{(weather.visibility / 1000).toFixed(1)} km</span>
            <span className="text-white/40 font-light text-xs">visibility</span>
          </div>
        </div>
        <div className="h-6 w-px bg-white/10 hidden md:block" />
        <div className="flex items-center gap-6 text-sm text-white/50 font-light">
          <span>H: {Math.round(celsiusToFahrenheit(weather.main.temp_max))}°</span>
          <span>L: {Math.round(celsiusToFahrenheit(weather.main.temp_min))}°</span>
          <span>Sunrise {formatTime(weather.sys.sunrise)}</span>
          <span>Sunset {formatTime(weather.sys.sunset)}</span>
        </div>
      </div>
    </div>
  );
}

