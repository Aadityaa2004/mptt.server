import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/react"
import { AuthProvider } from "./AuthContext"

describe("AuthProvider", () => {
  it("renders children", () => {
    render(
      <AuthProvider>
        <span>Child</span>
      </AuthProvider>
    )
    expect(screen.getByText("Child")).toBeInTheDocument()
  })
})
