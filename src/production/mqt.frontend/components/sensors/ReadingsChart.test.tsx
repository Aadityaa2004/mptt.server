import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/react"
import { ReadingsChart } from "./ReadingsChart"

describe("ReadingsChart", () => {
  it("renders with empty readings", () => {
    render(<ReadingsChart readings={[]} timeRange="1d" />)
    expect(document.body).toBeTruthy()
  })

  it("renders with readings", () => {
    const readings = [{
      ts: new Date().toISOString(),
      payload: {
        sensors: {
          temperature: { value: 10, unit: "F" },
          level: { value: 5, unit: "%" },
        },
        battery_percentage: 90,
      },
    }] as any
    const { container } = render(<ReadingsChart readings={readings} timeRange="1h" />)
    expect(container).toBeTruthy()
  })
})
