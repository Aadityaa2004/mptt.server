import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import { DeviceMap } from "./DeviceMap"

vi.mock("react-map-gl/maplibre", () => ({
  default: ({ children }: { children: React.ReactNode }) => <div data-testid="map">{children}</div>,
  Marker: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  Popup: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))
vi.mock("@/hooks/usePiPreferences", () => ({
  usePiPreferences: () => ({
    getPreference: () => null,
    initializePreferences: vi.fn(),
    loadColorsFromBackend: vi.fn(),
  }),
  colorToGradient: (c: string) => c,
}))

describe("DeviceMap", () => {
  it("renders", () => {
    render(
      <DeviceMap
        devices={[]}
        onDeviceAdd={vi.fn()}
        onDeviceClick={vi.fn()}
      />
    )
    expect(screen.getByTestId("map")).toBeInTheDocument()
  })
})
