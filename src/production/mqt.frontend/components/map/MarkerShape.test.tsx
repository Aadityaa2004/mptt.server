import { describe, it, expect } from "vitest"
import { render } from "@testing-library/react"
import { MarkerShapeComponent } from "./MarkerShape"

describe("MarkerShape", () => {
  it("renders", () => {
    const { container } = render(<MarkerShapeComponent gradient="from-blue-500 to-purple-500" />)
    expect(container.firstChild).toBeInTheDocument()
  })
})
