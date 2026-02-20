import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import { LocationInput } from "./LocationInput"

describe("LocationInput", () => {
  it("renders search input with placeholder", () => {
    render(<LocationInput onLocationSubmit={vi.fn()} />)
    const input = screen.getByPlaceholderText(/search for a city/i)
    expect(input, "LocationInput should render search input").toBeInTheDocument()
  })

  it("renders Set Location button", () => {
    render(<LocationInput onLocationSubmit={vi.fn()} />)
    const btn = screen.getByRole("button", { name: /set location/i })
    expect(btn, "LocationInput should render Set Location button").toBeInTheDocument()
  })
})
