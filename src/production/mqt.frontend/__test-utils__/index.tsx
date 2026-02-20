/**
 * Test utilities for frontend tests.
 * Use renderWithProviders for components that need auth/router context.
 * Use logTestFailure for descriptive error messages when assertions fail.
 */
import React, { ReactElement } from "react"
import { render, RenderOptions } from "@testing-library/react"

/** Log helper - use in expect().toBeTruthy() failures to pinpoint issues */
export function logTestFailure(componentName: string, detail: string) {
  return `[TEST FAIL] ${componentName}: ${detail}`
}

/** Wrapper for components needing common providers - extend as needed */
export function renderWithProviders(
  ui: ReactElement,
  options?: Omit<RenderOptions, "wrapper">
) {
  return render(ui, options)
}
