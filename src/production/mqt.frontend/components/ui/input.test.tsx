import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/react"
import { Input } from "./input"

describe("Input", () => {
  it("renders", () => {
    render(<Input placeholder="Enter text" />)
    expect(screen.getByPlaceholderText("Enter text")).toBeInTheDocument()
  })

  it("forwards ref and props", () => {
    render(<Input data-testid="input" type="email" />)
    const el = screen.getByTestId("input")
    expect(el).toHaveAttribute("type", "email")
  })
})
