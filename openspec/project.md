# Project Context

## Purpose
Go utilities for concurrency patterns such as backoff, cancellation-aware helpers, notifications, and token bucket rate limiting used across Plan42 services.

## Tech Stack
- Go 1.24
- Standard library primitives layered with custom helpers

## Project Conventions

### Code Style
Use gofmt formatting and idiomatic naming. `make lint` runs golangci-lint; keep exported APIs small and documented for reuse across services.

### Architecture Patterns
Single Go module organized by concern (e.g., backoff strategies, rate limiting, context utilities). Utilities are lightweight and avoid external dependencies.

### Testing Strategy
`go test ./...` covers concurrency behaviors; ensure tests handle timing deterministically where possible. Run linting before tagging releases.

### Git Workflow
Feature branches merged via PR. Tag releases with `make tag` when publishing module updates.

## Domain Context
Utilities are shared across multiple services, so API stability matters. Functions often assume contexts for cancellation and provide backoff defaults suitable for network retries.

## Important Constraints
- Avoid introducing heavy dependencies; keep the module minimal.
- Concurrency helpers must remain race-free and safe for use in high-volume services.

## External Dependencies
None; pure Go utilities.
