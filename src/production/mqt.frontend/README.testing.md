# Frontend Testing Guide

## Running Tests

| Command | Description |
|---------|-------------|
| `npm test` | Run all tests (parallel) |
| `npm run test:verbose` | Run with verbose reporter (shows each test name) |
| `npm run test:ci` | Run sequentially with verbose output; increases Node heap (4GB) to avoid OOM on large suites |

## Test Structure

- **Colocated tests:** Each `.tsx` file has a corresponding `.test.tsx` file in the same directory.
- **Setup:** `vitest.setup.ts` mocks `next/navigation`, `next/font/google`, canvas, ResizeObserver, and global fetch.
- **Assertions:** Tests use descriptive failure messages: `expect(el, "Component should render X").toBeInTheDocument()`.

## Debugging Failures

1. **Verbose output:** Run `npm run test:verbose` to see each test name as it runs.
2. **Single file:** `npx vitest run path/to/Component.test.tsx` to run one test file.
3. **Heap errors:** If you see "JavaScript heap out of memory", use `npm run test:ci` or `NODE_OPTIONS='--max-old-space-size=4096' npm test`.
4. **Failure messages:** Vitest shows accessible roles and DOM structure when queries fail (e.g. "Unable to find an accessible element with role 'button'").

## Mocks

- **AuthContext:** Pages and components that use `useAuth` or `useRequireAuth` mock them per-file.
- **API services:** `sensorService`, `weatherService`, `adminService`, `deviceLocationService` are mocked in page tests.
- **next/link, next/image:** Mocked globally in setup; some tests add per-file overrides.
