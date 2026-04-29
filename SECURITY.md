# Security Policy

## Reporting a Vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Use GitHub's private security disclosure:

1. Go to the [Security tab](../../security) of this repository
2. Click **"Report a vulnerability"**
3. Describe the issue with as much detail as possible

We will acknowledge your report within 72 hours and keep you updated as we work on a fix.

## Scope

In scope:
- Authentication bypass
- Unauthorized access to another org's data
- Remote code execution via the webhook handler or admin panel
- SQL injection via the raw query endpoint

Out of scope:
- Issues that require physical access to the server
- Denial of service via high webhook volume (by design, the server accepts all incoming webhooks)
- Issues in third-party dependencies (report those upstream)
