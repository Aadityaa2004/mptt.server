import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/react"
import { CurrentWeather } from "./CurrentWeather"
import type { CurrentWeather as CurrentWeatherType } from "@/types/weather"

const mockWeather: CurrentWeatherType = {
  coord: { lon: -79.38, lat: 43.65 },
  base: "stations",
  main: {
    temp: 15,
    feels_like: 14,
    temp_min: 12,
    temp_max: 18,
    pressure: 1015,
    humidity: 70,
  },
  visibility: 10000,
  wind: { speed: 5, deg: 180 },
  clouds: { all: 20 },
  dt: 1234567890,
  sys: { country: "CA", sunrise: 1234567890, sunset: 1234567890 },
  timezone: -18000,
  id: 6167865,
  name: "Toronto",
  cod: 200,
  weather: [{ id: 801, main: "Clouds", description: "few clouds", icon: "02d" }],
}

describe("CurrentWeather", () => {
  it("renders location and temp", () => {
    render(<CurrentWeather weather={mockWeather} />)
    expect(screen.getByText(/Toronto/)).toBeInTheDocument()
    expect(screen.getByText(/59°F/)).toBeInTheDocument()
  })
})
