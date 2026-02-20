import { describe, it, expect } from "vitest"
import { render } from "@testing-library/react"
import { DottedGlowBackground } from "./dotted-glow-background"

describe("DottedGlowBackground", () => {
  it("renders canvas element", () => {
    const { container } = render(<DottedGlowBackground />)
    const canvas = container.querySelector("canvas")
    expect(canvas, "DottedGlowBackground should render a canvas element").toBeInTheDocument()
  })
})
