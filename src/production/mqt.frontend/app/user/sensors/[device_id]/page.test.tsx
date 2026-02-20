import { describe, it, expect, vi } from "vitest"
import { render } from "@testing-library/react"
import DevicePage from "./page"

vi.mock("@/contexts/AuthContext", () => ({ useAuth: () => ({ user: { role: "user" }, isAuthenticated: true, logout: vi.fn() }) }))
vi.mock("@/hooks/useRequireAuth", () => ({ useRequireAuth: () => ({ user: { role: "user" }, isLoading: false }) }))
vi.mock("next/link", () => ({ default: ({ children }: { children: React.ReactNode }) => <a>{children}</a> }))
vi.mock("next/image", () => ({ default: (p: { alt: string }) => <img alt={p.alt} /> }))
vi.mock("@/services/api/sensorService", () => ({
  sensorService: {
    getPis: vi.fn().mockResolvedValue({ items: [{ pi_id: "pi-1" }] }),
    getDevices: vi.fn().mockResolvedValue({ items: [{ device_id: "dev-1", pi_id: "pi-1" }] }),
    getDeviceReadings: vi.fn().mockResolvedValue({ readings: [], next_page_token: null }),
    getLatestDeviceReading: vi.fn().mockResolvedValue(null),
  },
}))

describe("Device page", () => {
  it("renders", () => {
    const { container } = render(<DevicePage />)
    expect(container).toBeTruthy()
  })
})
