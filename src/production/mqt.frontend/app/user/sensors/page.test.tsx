import { describe, it, expect, vi } from "vitest"
import { render } from "@testing-library/react"
import SensorsPage from "./page"

vi.mock("@/contexts/AuthContext", () => ({ useAuth: () => ({ user: { role: "user" }, isAuthenticated: true, logout: vi.fn() }) }))
vi.mock("next/link", () => ({ default: ({ children }: { children: React.ReactNode }) => <a>{children}</a> }))
vi.mock("next/image", () => ({ default: (p: { alt: string }) => <img alt={p.alt} /> }))
vi.mock("@/hooks/useRequireAuth", () => ({ useRequireAuth: () => ({ user: { role: "user" }, isLoading: false }) }))
vi.mock("@/services/api/sensorService", () => ({ sensorService: { getPis: vi.fn().mockResolvedValue({ items: [] }) } }))

describe("Sensors page", () => {
  it("renders", () => {
    const { container } = render(<SensorsPage />)
    expect(container).toBeTruthy()
  })
})
