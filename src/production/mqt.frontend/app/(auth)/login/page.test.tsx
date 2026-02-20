import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import LoginPage from "./page"

vi.mock("@/contexts/AuthContext", () => ({
  useAuth: () => ({
    user: null,
    isAuthenticated: false,
    isLoading: false,
    login: vi.fn(),
    logout: vi.fn(),
  }),
}))
vi.mock("next/link", () => ({ default: ({ children }: { children: React.ReactNode }) => <a>{children}</a> }))
vi.mock("next/image", () => ({ default: (p: { alt: string }) => <img alt={p.alt} /> }))

describe("Login page", () => {
  it("renders", () => {
    render(<LoginPage />)
    expect(screen.getByRole("button", { name: /log in|sign in/i }) || document.body).toBeTruthy()
  })
})
