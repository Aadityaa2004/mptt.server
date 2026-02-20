import { describe, it, expect, vi } from "vitest"
import { render } from "@testing-library/react"
import AdminDashboardPage from "./page"

vi.mock("@/contexts/AuthContext", () => ({ useAuth: () => ({ user: { role: "admin" }, isAuthenticated: true, logout: vi.fn() }) }))
vi.mock("next/link", () => ({ default: ({ children }: { children: React.ReactNode }) => <a>{children}</a> }))
vi.mock("next/image", () => ({ default: (p: { alt: string }) => <img alt={p.alt} /> }))
vi.mock("@/hooks/useRequireAuth", () => ({ useRequireAuth: () => ({ user: { role: "admin" }, isLoading: false }) }))
vi.mock("@/services/api/adminService", () => ({
  adminService: {
    getAllUsers: vi.fn().mockResolvedValue({ users: [] }),
    getAllPis: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, page_size: 1000 }),
    getDevices: vi.fn().mockResolvedValue({ items: [], total: 0 }),
    getReadings: vi.fn().mockResolvedValue({ readings: [], next_page_token: null }),
    getDeviceReadings: vi.fn().mockResolvedValue({ readings: [] }),
    getApiHealth: vi.fn().mockResolvedValue({ status: "ok" }),
    getSummaryStats: vi.fn().mockResolvedValue(null),
  },
}))

describe("Admin dashboard page", () => {
  it("renders", () => {
    const { container } = render(<AdminDashboardPage />)
    expect(container).toBeTruthy()
  })
})
