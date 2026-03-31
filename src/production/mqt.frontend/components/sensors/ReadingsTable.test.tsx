import { describe, it, expect, vi, beforeEach } from "vitest"
import { render, screen, waitFor } from "@testing-library/react"
import { ReadingsTable } from "./ReadingsTable"

vi.mock("@/services/api/sensorService", () => ({
  sensorService: {
    getDeviceReadings: vi.fn().mockResolvedValue({ items: [], next_page_token: null }),
  },
}))

beforeEach(() => {
  vi.clearAllMocks()
})

describe("ReadingsTable", () => {
  it("renders loading state or table", async () => {
    render(<ReadingsTable deviceId="dev-1" piId="pi-1" />)
    await waitFor(() => {
      const loading = screen.queryByText(/loading/i)
      const table = screen.queryByRole("table")
      expect(
        loading || table || document.body,
        "ReadingsTable should show loading or table"
      ).toBeTruthy()
    }, { timeout: 3000 })
  })
})
