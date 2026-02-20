import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import { DeviceCard } from "./DeviceCard"

vi.mock("@/services/api/sensorService", () => ({
  sensorService: {
    getLatestDeviceReading: vi.fn().mockResolvedValue(null),
  },
}))

describe("DeviceCard", () => {
  it("renders device id", () => {
    render(<DeviceCard device={{ pi_id: "pi-1", device_id: "dev-1" }} />)
    expect(
      screen.getByText("dev-1", { exact: false }),
      "DeviceCard should display device_id"
    ).toBeInTheDocument()
  })
})
