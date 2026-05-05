# Security Policy

## Supported versions

shipnote is pre-1.0. Only the latest minor release receives security fixes.

| Version | Supported |
| ------- | --------- |
| latest  | ✅        |
| older   | ❌        |

## Reporting a vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Please report privately via [GitHub's private vulnerability reporting](https://github.com/DesmondSanctity/shipnote/security/advisories/new).

Include:

- A description of the issue and its impact
- Reproduction steps or a proof of concept
- Affected version(s)
- Any suggested mitigations

## Process

1. We acknowledge receipt within 5 working days.
2. We assess severity and produce a fix on a private branch.
3. We coordinate a release date with the reporter.
4. We publish a GitHub Security Advisory and patch release simultaneously.
5. Reporters are credited in the advisory unless they request otherwise.

## Scope

In scope:

- Code execution via crafted repository content (commit messages, PR bodies, config files)
- Path traversal in any file-writing code path
- Credential leakage in logs or generated output
- Supply-chain integrity (release artifacts, Docker images, npm wrapper)

Out of scope:

- Vulnerabilities in upstream dependencies (please report to upstream first)
- Issues requiring an attacker who already has shell access to the host
- Rate-limit / denial-of-wallet on third-party APIs (GitHub, AI providers)

## Threat model summary

shipnote runs locally or in CI with read access to a git repository and (optionally) a GitHub token. It does not store credentials, does not phone home, does not auto-update. AI features are off by default in CI and never receive credentials by default.
