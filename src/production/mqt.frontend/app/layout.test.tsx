import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/react"
import RootLayout from "./layout"

describe("RootLayout", () => {
  it("renders children", () => {
    render(
      <RootLayout>
        <div>Test child</div>
      </RootLayout>
    )
    expect(screen.getByText("Test child")).toBeInTheDocument()
  })
})
