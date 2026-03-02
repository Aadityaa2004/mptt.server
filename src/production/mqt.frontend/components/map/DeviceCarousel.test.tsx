import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import { DeviceCarousel } from "./DeviceCarousel"
import type { Device } from "@/types/device"

vi.mock("@/services/api/sensorService", () => ({
  sensorService: { getLatestDeviceReading: vi.fn().mockResolvedValue(null) },
}))
vi.mock("@/services/api/deviceLocationService", () => ({
  deviceLocationService: { deleteLocation: vi.fn().mockResolvedValue(undefined) },
}))

const mockDevice: Device = {
  id: "1",
  pi_id: "pi-1",
  device_id: "dev-1",
  latitude: 0,
  longitude: 0,
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
}

describe("DeviceCarousel", () => {
  it("renders device heading and content", () => {
    render(
      <DeviceCarousel
        devices={[mockDevice]}
        currentIndex={0}
        onClose={vi.fn()}
        onNavigate={vi.fn()}
      />
    )
    expect(
      screen.getByText("dev-1"),
      "DeviceCarousel should display device_id"
    ).toBeInTheDocument()
    expect(
      screen.getByText(/pi-1/),
      "DeviceCarousel should display pi_id"
    ).toBeInTheDocument()
  })
})
