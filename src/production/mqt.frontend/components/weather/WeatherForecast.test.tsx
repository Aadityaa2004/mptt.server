import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/react"
import { WeatherForecast } from "./WeatherForecast"

const mockForecast = {
  list: [
    {
      dt: Math.floor(Date.now() / 1000),
      main: { temp: 15, temp_max: 18, temp_min: 12, humidity: 70 },
      weather: [{ icon: "02d", description: "clouds" }],
      pop: 0.1,
      wind: { speed: 5 },
      rain: undefined as any,
      snow: undefined as any,
    },
  ],
} as any

describe("WeatherForecast", () => {
  it("renders with forecast", () => {
    render(<WeatherForecast forecast={mockForecast} />)
    expect(document.body).toBeTruthy()
  })
})
