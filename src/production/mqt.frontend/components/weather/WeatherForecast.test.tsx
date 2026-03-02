import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/react"
import { WeatherForecast } from "./WeatherForecast"
import { WeatherForecast as WeatherForecastType } from "@/types/weather"

const mockForecast: WeatherForecastType = {
  cod: "200",
  message: 0,
  cnt: 1,
  list: [
    {
      dt: Math.floor(Date.now() / 1000),
      main: {
        temp: 15,
        feels_like: 15,
        temp_max: 18,
        temp_min: 12,
        pressure: 1013,
        humidity: 70,
        temp_kf: 0,
      },
      weather: [
        {
          id: 1,
          main: "Clouds",
          description: "clouds",
          icon: "02d",
        },
      ],
      clouds: { all: 75 },
      wind: { speed: 5, deg: 180 },
      visibility: 10000,
      pop: 0.1,
      sys: { pod: "d" },
      dt_txt: new Date().toISOString(),
    },
  ],
  city: {
    id: 1,
    name: "Test City",
    coord: { lon: 0, lat: 0 },
    country: "US",
    population: 0,
    timezone: 0,
    sunrise: Math.floor(Date.now() / 1000),
    sunset: Math.floor(Date.now() / 1000),
  },
}

describe("WeatherForecast", () => {
  it("renders with forecast", () => {
    render(<WeatherForecast forecast={mockForecast} />)
    expect(document.body).toBeTruthy()
  })
})
