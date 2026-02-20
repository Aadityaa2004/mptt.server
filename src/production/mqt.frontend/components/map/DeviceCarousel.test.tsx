import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import { DeviceCarousel } from "./DeviceCarousel"

vi.mock("@/services/api/sensorService", () => ({
  sensorService: { getLatestDeviceReading: vi.fn().mockResolvedValue(null) },
}))
vi.mock("@/services/api/deviceLocationService", () => ({
  deviceLocationService: { deleteDeviceLocation: vi.fn().mockResolvedValue(undefined) },
}))

const mockDevice = { id: "1", pi_id: "pi-1", device_id: "dev-1", latitude: 0, longitude: 0 } as any

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
      screen.getByRole("heading", { name: /device/i }),
      "DeviceCarousel should display Device heading"
    ).toBeInTheDocument()
    expect(
      screen.getByText("dev-1"),
      "DeviceCarousel should display device_id"
    ).toBeInTheDocument()
  })
})
