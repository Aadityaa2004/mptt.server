# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in this project, please report it responsibly:

- **Do not** open a public GitHub/GitLab issue for security-sensitive bugs.
- Email the maintainers privately or report via your organization’s secure channel (e.g. private issue, security contact).
- Include a clear description, steps to reproduce, and impact if possible.

We aim to acknowledge reports within **5 business days** and to provide an initial assessment or fix timeline shortly after.

## Expectations

- **No sensitive data in the repo:** Never commit `.env.production`, API keys, passwords, or tokens. Use environment variables and secrets managers in deployment.
- **Dependencies:** We run `go mod verify`, `govulncheck`, and `npm audit` in CI. Please keep dependencies updated and report any known vulnerabilities you notice.

## Supported Versions

Security updates are applied to the current main branch and the latest release tag. Older versions may not receive patches; please upgrade when possible.
