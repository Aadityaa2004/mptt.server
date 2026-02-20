import { describe, it, expect, vi } from "vitest"
import { render } from "@testing-library/react"
import SettingsPage from "./page"

vi.mock("@/contexts/AuthContext", () => ({ useAuth: () => ({ user: { role: "user" }, isAuthenticated: true, logout: vi.fn() }) }))
vi.mock("next/link", () => ({ default: ({ children }: { children: React.ReactNode }) => <a>{children}</a> }))
vi.mock("next/image", () => ({ default: (p: { alt: string }) => <img alt={p.alt} /> }))

describe("Settings page", () => {
  it("renders", () => {
    const { container } = render(<SettingsPage />)
    expect(container).toBeTruthy()
  })
})
