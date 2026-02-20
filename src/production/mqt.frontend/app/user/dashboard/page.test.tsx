import { describe, it, expect, vi } from "vitest"
import { render } from "@testing-library/react"
import DashboardPage from "./page"

vi.mock("@/contexts/AuthContext", () => ({ useAuth: () => ({ user: { role: "user" }, isAuthenticated: true, logout: vi.fn() }) }))
vi.mock("@/hooks/useRequireAuth", () => ({ useRequireAuth: () => ({ user: { role: "user" }, isLoading: false }) }))
vi.mock("next/link", () => ({ default: ({ children }: { children: React.ReactNode }) => <a>{children}</a> }))
vi.mock("next/image", () => ({ default: (p: { alt: string }) => <img alt={p.alt} /> }))
vi.mock("@/services/api/sensorService", () => ({ sensorService: { getPis: vi.fn().mockResolvedValue({ items: [] }) } }))
vi.mock("@/services/api/weatherService", () => ({
  weatherService: {
    getProfile: vi.fn().mockResolvedValue(null),
    getCurrentWeather: vi.fn().mockResolvedValue(null),
    updateLocation: vi.fn().mockResolvedValue(undefined),
  },
}))
vi.mock("@/services/api/deviceLocationService", () => ({
  deviceLocationService: {
    getAllLocations: vi.fn().mockResolvedValue([]),
    convertToDevice: vi.fn((l: any, id: string) => ({ id, ...l })),
  },
}))

describe("User dashboard page", () => {
  it("renders", () => {
    const { container } = render(<DashboardPage />)
    expect(container).toBeTruthy()
  })
})
