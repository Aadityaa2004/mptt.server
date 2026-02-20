import { describe, it, expect, vi } from "vitest"
import { render } from "@testing-library/react"
import MapPage from "./page"

vi.mock("@/contexts/AuthContext", () => ({ useAuth: () => ({ user: null, isAuthenticated: false, logout: vi.fn() }) }))
vi.mock("next/link", () => ({ default: ({ children }: { children: React.ReactNode }) => <a>{children}</a> }))
vi.mock("next/image", () => ({ default: (p: { alt: string }) => <img alt={p.alt} /> }))
vi.mock("react-map-gl/maplibre", () => ({
  default: ({ children }: { children: React.ReactNode }) => <div data-testid="map">{children}</div>,
  Marker: () => null,
  Popup: () => null,
}))
vi.mock("@/hooks/usePiPreferences", () => ({
  usePiPreferences: () => ({ getPreference: () => null, initializePreferences: vi.fn(), loadColorsFromBackend: vi.fn() }),
  colorToGradient: (c: string) => c,
}))
vi.mock("@/services/api/sensorService", () => ({ sensorService: { getDevices: vi.fn().mockResolvedValue([]), getPis: vi.fn().mockResolvedValue([]) } }))
vi.mock("@/services/api/deviceLocationService", () => ({ deviceLocationService: { getDeviceLocations: vi.fn().mockResolvedValue([]) } }))

describe("Map page", () => {
  it("renders", () => {
    const { container } = render(<MapPage />)
    expect(container).toBeTruthy()
  })
})
