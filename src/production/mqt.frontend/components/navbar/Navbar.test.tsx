import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import Navbar from "./Navbar"

vi.mock("@/contexts/AuthContext", () => ({
  useAuth: () => ({
    user: null,
    isAuthenticated: false,
    logout: vi.fn(),
  }),
}))
vi.mock("next/link", () => ({ default: ({ children, href }: { children: React.ReactNode; href: string }) => <a href={href}>{children}</a> }))
vi.mock("next/image", () => ({ default: (props: { alt: string }) => <img alt={props.alt} /> }))

describe("Navbar", () => {
  it("renders", () => {
    render(<Navbar />)
    expect(screen.getByRole("navigation")).toBeInTheDocument()
  })
})
