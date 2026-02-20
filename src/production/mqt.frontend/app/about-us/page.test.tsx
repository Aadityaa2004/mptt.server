import { describe, it, expect, vi, beforeEach } from "vitest"
import { render } from "@testing-library/react"
import AboutUsPage from "./page"

vi.mock("@/contexts/AuthContext", () => ({ useAuth: () => ({ user: null, isAuthenticated: false, logout: vi.fn() }) }))
vi.mock("next/link", () => ({ default: ({ children }: { children: React.ReactNode }) => <a>{children}</a> }))
vi.mock("next/image", () => ({ default: (p: { alt: string }) => <img alt={p.alt} /> }))

describe("AboutUs page", () => {
  it("renders", () => {
    const { container } = render(<AboutUsPage />)
    expect(container).toBeTruthy()
  })
})
