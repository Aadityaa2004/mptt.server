import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/react"
import { CurrentWeather } from "./CurrentWeather"

const mockWeather = {
  name: "Toronto",
  main: {
    temp: 15,
    feels_like: 14,
    pressure: 1015,
    humidity: 70,
  },
  weather: [{ main: "Clouds", description: "few clouds", icon: "02d" }],
  wind: { speed: 5 },
  visibility: 10000,
  sys: { sunrise: 1234567890, sunset: 1234567890 },
} as any

describe("CurrentWeather", () => {
  it("renders location and temp", () => {
    render(<CurrentWeather weather={mockWeather} />)
    expect(screen.getByText("Toronto")).toBeInTheDocument()
    expect(screen.getByText(/59°F/)).toBeInTheDocument()
  })
})
