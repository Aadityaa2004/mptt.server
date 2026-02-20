import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/react"
import { ErrorBoundary } from "./ErrorBoundary"

const Throw = () => {
  throw new Error("Test error")
}

describe("ErrorBoundary", () => {
  it("renders children when no error", () => {
    render(
      <ErrorBoundary>
        <span>Child content</span>
      </ErrorBoundary>
    )
    expect(screen.getByText("Child content")).toBeInTheDocument()
  })

  it("renders fallback when child throws", () => {
    render(
      <ErrorBoundary fallback={<div>Fallback</div>}>
        <Throw />
      </ErrorBoundary>
    )
    expect(screen.getByText("Fallback")).toBeInTheDocument()
  })
})
