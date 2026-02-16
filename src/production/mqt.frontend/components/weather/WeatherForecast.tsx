"use client";

import { useState, useMemo } from "react";
import { WeatherForecast as WeatherForecastType } from "@/types/weather";
import {
  ComposedChart,
  Line,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  AreaChart,
  Area,
  Legend,
  ReferenceLine,
} from "recharts";
import { Calendar, Droplet, CloudRain, Snowflake, Thermometer } from "lucide-react";

interface WeatherForecastProps {
  forecast: WeatherForecastType;
}

// Convert Celsius to Fahrenheit
const celsiusToFahrenheit = (celsius: number): number => {
  return (celsius * 9) / 5 + 32;
};

export function WeatherForecast({ forecast }: WeatherForecastProps) {
  const [selectedDays, setSelectedDays] = useState<number>(5);

  // Calculate how many data points to show based on selected days
  // OpenWeatherMap provides 3-hour intervals, so ~8 per day
  const dataPointsPerDay = 8;
  const maxDataPoints = selectedDays * dataPointsPerDay;

  // Prepare hourly data points for the chart (bumpy graph with all data points)
  const chartData = useMemo(() => {
    return forecast.list.slice(0, maxDataPoints).map((item) => {
      const date = new Date(item.dt * 1000);
      const dayName = date.toLocaleDateString("en-US", { weekday: "short" });
      const time = date.toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit" });
      
      return {
        time: `${dayName} ${time}`,
        day: dayName,
        date: date.toLocaleDateString("en-US", { month: "short", day: "numeric" }),
        hour: date.getHours(),
        temp: Math.round(celsiusToFahrenheit(item.main.temp)),
        tempMax: Math.round(celsiusToFahrenheit(item.main.temp_max)),
        tempMin: Math.round(celsiusToFahrenheit(item.main.temp_min)),
        humidity: item.main.humidity,
        pop: Math.round(item.pop * 100),
        rain: item.rain?.["3h"] || 0,
        snow: item.snow?.["3h"] || 0,
        windSpeed: item.wind.speed,
        icon: item.weather[0].icon,
        description: item.weather[0].description,
      };
    });
  }, [forecast.list, maxDataPoints]);

  // Calculate average/sum values for display
  const avgValues = useMemo(() => {
    const data = chartData;
    if (data.length === 0) return { temp: 0, humidity: 0, pop: 0, rain: 0, snow: 0 };
    
    const avgTemp = data.reduce((sum, d) => sum + d.temp, 0) / data.length;
    const avgHumidity = data.reduce((sum, d) => sum + d.humidity, 0) / data.length;
    const maxPop = Math.max(...data.map((d) => d.pop));
    const totalRain = data.reduce((sum, d) => sum + d.rain, 0);
    const totalSnow = data.reduce((sum, d) => sum + d.snow, 0);
    
    return {
      temp: Math.round(avgTemp),
      humidity: Math.round(avgHumidity),
      pop: maxPop,
      rain: Math.round(totalRain * 10) / 10, // Round to 1 decimal
      snow: Math.round(totalSnow * 10) / 10,
    };
  }, [chartData]);

  const TempTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      const data = payload[0].payload;
      return (
        <div className="bg-black/90 border border-orange-500/30 rounded-lg p-3 shadow-lg min-w-[160px]">
          <p className="text-white/80 text-xs mb-2 font-medium">{data.time}</p>
          <p className="text-orange-300 font-light text-lg">
            {data.temp}°F
          </p>
          <p className="text-white/60 text-xs mt-1">
            {data.tempMax}°F / {data.tempMin}°F
          </p>
        </div>
      );
    }
    return null;
  };

  const PrecipitationTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      const data = payload[0].payload;
      return (
        <div className="bg-black/90 border border-blue-500/30 rounded-lg p-3 shadow-lg min-w-[180px]">
          <p className="text-white/80 text-xs mb-2 font-medium">{data.time}</p>
          <div className="space-y-1 text-xs">
            <p className="text-white/70">
              <span className="text-blue-300">Humidity:</span> {data.humidity}%
            </p>
            {data.rain > 0 && (
              <p className="text-white/70">
                <span className="text-blue-400">Rainfall:</span> {data.rain.toFixed(1)} mm
              </p>
            )}
            {data.snow > 0 && (
              <p className="text-white/70">
                <span className="text-cyan-200">Snowfall:</span> {data.snow.toFixed(1)} mm
              </p>
            )}
          </div>
        </div>
      );
    }
    return null;
  };

  return (
    <div className="space-y-6 select-none">
      {/* Filter Controls */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Calendar className="h-5 w-5 text-white/60" />
          <h3 className="text-xl font-light">Weather Forecast</h3>
        </div>
        <div className="flex items-center gap-2">
          {[1, 2, 3, 4, 5].map((days) => (
            <button
              key={days}
              onClick={() => setSelectedDays(days)}
              className={`px-3 py-1.5 text-xs font-light rounded-md transition-all ${
                selectedDays === days
                  ? "bg-gradient-to-r from-orange-500/20 to-orange-600/20 border border-orange-500/40 text-orange-300"
                  : "bg-white/5 border border-white/10 text-white/60 hover:bg-white/10 hover:text-white/80"
              }`}
            >
              {days} {days === 1 ? "Day" : "Days"}
            </button>
          ))}
        </div>
      </div>

      {/* Two Side-by-Side Graphs */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Temperature Graph */}
        <div className="border border-white/10 rounded-lg p-6 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm">
          <div className="flex items-center gap-2 mb-6">
            <Thermometer className="h-5 w-5 text-orange-300" />
            <h3 className="text-lg font-light text-orange-300">Temperature</h3>
          </div>
          
          {/* Temperature Summary */}
          <div className="flex items-center gap-3 mb-6 pb-4 border-b border-white/10">
            <div className="w-12 h-12 rounded-full bg-gradient-to-br from-orange-500/30 to-orange-600/30 border border-orange-500/40 flex items-center justify-center">
              <span className="text-orange-300 text-xl font-light">{avgValues.temp}°</span>
            </div>
            <div>
              <div className="text-xs text-white/60 font-light">Average Temperature</div>
              <div className="text-base text-white/90 font-light">{avgValues.temp}°F</div>
            </div>
          </div>

          <ResponsiveContainer width="100%" height={350}>
            <AreaChart data={chartData} margin={{ top: 10, right: 20, left: 0, bottom: 60 }}>
              <defs>
                <linearGradient id="tempGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="rgba(251, 146, 60, 0.4)" stopOpacity={0.6} />
                  <stop offset="50%" stopColor="rgba(249, 115, 22, 0.3)" stopOpacity={0.4} />
                  <stop offset="100%" stopColor="rgba(234, 88, 12, 0.1)" stopOpacity={0.1} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="rgba(255, 255, 255, 0.1)" />
              <XAxis
                dataKey="time"
                stroke="rgba(255, 255, 255, 0.6)"
                style={{ fontSize: "11px" }}
                angle={-45}
                textAnchor="end"
                height={80}
              />
              <YAxis
                stroke="rgba(255, 255, 255, 0.6)"
                style={{ fontSize: "12px" }}
                label={{ value: "°F", angle: -90, position: "insideLeft", fill: "rgba(255, 255, 255, 0.6)" }}
              />
              <Tooltip content={<TempTooltip />} />
              <Area
                type="monotone"
                dataKey="temp"
                stroke="rgba(251, 146, 60, 1)"
                strokeWidth={3}
                fillOpacity={1}
                fill="url(#tempGradient)"
                name="Temperature"
                dot={{ fill: "rgba(251, 146, 60, 1)", r: 4 }}
                activeDot={{ r: 6, fill: "rgba(249, 115, 22, 1)" }}
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>

        {/* Humidity, Rainfall & Snow Graph */}
        <div className="border border-white/10 rounded-lg p-6 bg-gradient-to-br from-white/5 to-white/0 backdrop-blur-sm">
          <div className="flex items-center gap-2 mb-6">
            <Droplet className="h-5 w-5 text-blue-300" />
            <h3 className="text-lg font-light text-blue-300">Humidity & Precipitation</h3>
          </div>

          {/* Summary Stats */}
          <div className="flex items-center gap-4 mb-6 pb-4 border-b border-white/10 flex-wrap">
            <div className="flex items-center gap-2">
              <div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-500/20 to-blue-600/20 border border-blue-500/30 flex items-center justify-center">
                <Droplet className="h-4 w-4 text-blue-300" />
              </div>
              <div>
                <div className="text-xs text-white/60 font-light">Humidity</div>
                <div className="text-sm text-white/90 font-light">{avgValues.humidity}%</div>
              </div>
            </div>
            {avgValues.rain > 0 && (
              <div className="flex items-center gap-2">
                <div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-400/20 to-blue-500/20 border border-blue-400/30 flex items-center justify-center">
                  <CloudRain className="h-4 w-4 text-blue-400" />
                </div>
                <div>
                  <div className="text-xs text-white/60 font-light">Rainfall</div>
                  <div className="text-sm text-white/90 font-light">{avgValues.rain} mm</div>
                </div>
              </div>
            )}
            {avgValues.snow > 0 && (
              <div className="flex items-center gap-2">
                <div className="w-10 h-10 rounded-full bg-gradient-to-br from-cyan-300/20 to-cyan-400/20 border border-cyan-300/30 flex items-center justify-center">
                  <Snowflake className="h-4 w-4 text-cyan-200" />
                </div>
                <div>
                  <div className="text-xs text-white/60 font-light">Snowfall</div>
                  <div className="text-sm text-white/90 font-light">{avgValues.snow} mm</div>
                </div>
              </div>
            )}
          </div>

          <ResponsiveContainer width="100%" height={350}>
            <ComposedChart data={chartData} margin={{ top: 10, right: 20, left: 0, bottom: 60 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="rgba(255, 255, 255, 0.1)" />
              <XAxis
                dataKey="time"
                stroke="rgba(255, 255, 255, 0.6)"
                style={{ fontSize: "11px" }}
                angle={-45}
                textAnchor="end"
                height={80}
              />
              <YAxis
                yAxisId="humidity"
                stroke="rgba(255, 255, 255, 0.6)"
                style={{ fontSize: "12px" }}
                label={{ value: "Humidity %", angle: -90, position: "insideLeft", fill: "rgba(255, 255, 255, 0.6)" }}
                domain={[0, 100]}
              />
              <YAxis
                yAxisId="precipitation"
                orientation="right"
                stroke="rgba(255, 255, 255, 0.6)"
                style={{ fontSize: "12px" }}
                label={{ value: "Precipitation (mm)", angle: 90, position: "insideRight", fill: "rgba(255, 255, 255, 0.6)" }}
              />
              <Tooltip content={<PrecipitationTooltip />} />
              
              {/* Level Indicators */}
              <ReferenceLine yAxisId="humidity" y={50} stroke="rgba(59, 130, 246, 0.3)" strokeDasharray="2 2" label={{ value: "Comfortable", position: "insideTopRight", fill: "rgba(59, 130, 246, 0.6)", fontSize: 10 }} />
              <ReferenceLine yAxisId="humidity" y={70} stroke="rgba(59, 130, 246, 0.4)" strokeDasharray="2 2" label={{ value: "High", position: "insideTopRight", fill: "rgba(59, 130, 246, 0.7)", fontSize: 10 }} />
              
              {/* Humidity line */}
              <Line
                yAxisId="humidity"
                type="monotone"
                dataKey="humidity"
                stroke="rgba(59, 130, 246, 1)"
                strokeWidth={3}
                dot={{ fill: "rgba(59, 130, 246, 1)", r: 4 }}
                activeDot={{ r: 6 }}
                name="Humidity"
              />
              
              {/* Rainfall bars */}
              {chartData.some((d) => d.rain > 0) && (
                <Bar
                  yAxisId="precipitation"
                  dataKey="rain"
                  fill="rgba(96, 165, 250, 0.7)"
                  name="Rainfall"
                  radius={[3, 3, 0, 0]}
                />
              )}
              
              {/* Snowfall bars */}
              {chartData.some((d) => d.snow > 0) && (
                <Bar
                  yAxisId="precipitation"
                  dataKey="snow"
                  fill="rgba(165, 243, 252, 0.7)"
                  name="Snowfall"
                  radius={[3, 3, 0, 0]}
                />
              )}
              
              <Legend
                wrapperStyle={{ paddingTop: "20px" }}
                iconType="line"
                formatter={(value) => (
                  <span className="text-white/70 text-xs font-light">{value}</span>
                )}
              />
            </ComposedChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  );
}

