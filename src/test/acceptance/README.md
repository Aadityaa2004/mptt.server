# Acceptance / E2E tests

High-level flow tests (e.g. login, create Pi, list readings) run against the full stack.

**Prerequisites:** Stack running (e.g. `docker compose up` or RPi deployment). Set `TEST_API_URL` (and any auth test credentials) as needed.

**Run:** Add Go tests or Playwright/other E2E here; run manually or in CI as a separate step after the stack is up.
