import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/react"
import { Button } from "./button"

describe("Button", () => {
  it("renders children", () => {
    render(<Button>Click me</Button>)
    expect(screen.getByRole("button", { name: /click me/i })).toBeInTheDocument()
  })

  it("renders with variant and size", () => {
    render(<Button variant="outline" size="sm">Submit</Button>)
    expect(screen.getByRole("button", { name: /submit/i })).toBeInTheDocument()
  })
})
