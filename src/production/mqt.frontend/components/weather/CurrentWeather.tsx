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
    <div className="border border-white/10 rounded-lg p-4 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm select-none">
      <div className="flex items-start justify-between mb-4">
        <div>
          <h2 className="text-xl font-light mb-0.5">{weather.name}</h2>
          <p className="text-white/60 text-xs font-light capitalize">
            {mainCondition.description}
          </p>
        </div>
        <div className="flex items-center gap-2">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src={iconUrl} alt={mainCondition.main} className="w-12 h-12" />
          <div className="text-right">
            <div className="text-3xl font-light">{Math.round(celsiusToFahrenheit(weather.main.temp))}°F</div>
            <div className="text-xs text-white/60 font-light">
              Feels like {Math.round(celsiusToFahrenheit(weather.main.feels_like))}°F
            </div>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mt-4">
        <div className="flex items-center gap-2">
          <Gauge className="h-3.5 w-3.5 text-white/40" />
          <div>
            <div className="text-xs text-white/60 font-light">Pressure</div>
            <div className="text-xs font-light">{weather.main.pressure} hPa</div>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Droplets className="h-3.5 w-3.5 text-white/40" />
          <div>
            <div className="text-xs text-white/60 font-light">Humidity</div>
            <div className="text-xs font-light">{weather.main.humidity}%</div>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Wind className="h-3.5 w-3.5 text-white/40" />
          <div>
            <div className="text-xs text-white/60 font-light">Wind</div>
            <div className="text-xs font-light">{weather.wind.speed} m/s</div>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Eye className="h-3.5 w-3.5 text-white/40" />
          <div>
            <div className="text-xs text-white/60 font-light">Visibility</div>
            <div className="text-xs font-light">{(weather.visibility / 1000).toFixed(1)} km</div>
          </div>
        </div>
      </div>

      <div className="mt-4 pt-4 border-t border-white/10 grid grid-cols-2 gap-4">
        <div>
          <div className="text-xs text-white/60 font-light mb-0.5">Sunrise</div>
          <div className="text-xs font-light">{formatTime(weather.sys.sunrise)}</div>
        </div>
        <div>
          <div className="text-xs text-white/60 font-light mb-0.5">Sunset</div>
          <div className="text-xs font-light">{formatTime(weather.sys.sunset)}</div>
        </div>
      </div>

      <div className="mt-3 flex items-center gap-4 text-xs text-white/60 font-light">
        <div>
          High: <span className="text-white">{Math.round(celsiusToFahrenheit(weather.main.temp_max))}°F</span>
        </div>
        <div>
          Low: <span className="text-white">{Math.round(celsiusToFahrenheit(weather.main.temp_min))}°F</span>
        </div>
      </div>
    </div>
  );
}

