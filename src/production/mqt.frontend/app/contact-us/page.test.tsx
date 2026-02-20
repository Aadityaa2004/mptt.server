import { describe, it, expect, vi } from "vitest"
import { render } from "@testing-library/react"
import ContactUsPage from "./page"

vi.mock("@/contexts/AuthContext", () => ({ useAuth: () => ({ user: null, isAuthenticated: false, logout: vi.fn() }) }))
vi.mock("next/link", () => ({ default: ({ children }: { children: React.ReactNode }) => <a>{children}</a> }))
vi.mock("next/image", () => ({ default: (p: { alt: string }) => <img alt={p.alt} /> }))

describe("ContactUs page", () => {
  it("renders", () => {
    const { container } = render(<ContactUsPage />)
    expect(container).toBeTruthy()
  })
})
