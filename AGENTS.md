# AGENTS.md

Guidance for AI coding agents working in this repository.

## Project Overview

`http-little-toy` is a lightweight HTTP concurrency/load testing tool written in Go (module path: `github.com/leihenshang/http-little-toy`). It fires concurrent HTTP/1.1 or HTTP/2 requests against a target URL for a configured duration and reports statistics (RPS, transfer rate, response times, success/failure counts). It supports custom headers, request bodies, TLS/mTLS, and raw/JSON/CSV output.

- **Go version**: 1.22+ (see `go.mod`)
- **Dependencies**: only `golang.org/x/net/http2` — everything else is the standard library. Avoid adding new dependencies unless clearly justified.

## Repository Structure

| Path | Purpose |
|------|---------|
| `main.go` | CLI entry point: flag parsing, worker goroutines, HTTP client construction (`genHttpClient`), request execution (`doReq`), result aggregation |
| `data/` | Core data types: `ToyReq` (request config + `Validate`), `RequestStats` (statistics + `PrintStats`), global constants (`Version`, `AppName`) |
| `msg/` | Bilingual message localization (`en`/`zh`) via `ToyMsg` and `SetLocalize` |
| `utils/` | Small helpers (e.g. `CreateFile`) |
| `test-server/` | Standalone test HTTP/HTTPS server for manual testing (has its own `README.md`) |
| `rate_limit/` | Design notes / README only, not yet implemented |

## Common Commands

```bash
# Build
go build -o http-little-toy .

# Run all tests
go test ./...

# Run tests for a single package
go test ./data/ ./msg/ ./utils/

# Run with coverage
go test ./... -cover

# Vet / static checks
go vet ./...
gofmt -l .

# Run the tool directly (go run)
go run . -u http://example.com -t 10 -d 30
go run . -h          # help
go run . -v          # version

# Local manual testing against the bundled test server
go run ./test-server  # then point the tool at the server's address
```

Cross-platform builds use `CGO_ENABLED=0` with `GOOS`/`GOARCH` (see README "Cross-platform Compilation"). The binary must remain CGO-free.

## Code Style & Conventions

- Format all Go code with `gofmt` (standard, no exceptions); pass `go vet ./...` before finishing a change.
- Follow idiomatic Go: explicit error returns over panics; wrap errors with context using `fmt.Errorf("...: %w", err)`.
- Comments in this repo are mostly in Chinese; bilingual (zh/en) user-facing strings live in `msg/`. Match the style of the file you are editing.
- Package layout convention: business data/types in `data/`, user-facing messages in `msg/`, generic helpers in `utils/`. Keep `main.go` focused on orchestration.
- New CLI flags are registered in `initParameters()` in `main.go` and bound to fields on `data.ToyReq` (with sensible defaults and a description). Validation belongs in `ToyReq.Validate()`.
- When adding user-facing output, add both `Cn` and `En` variants to `msg/` rather than hardcoding strings.
- Bump `data.Version` in `data/global.go` only when the change warrants a release.

## Testing Conventions

- Tests are colocated (`*_test.go`) and use only the standard `testing` package (table-driven where natural); no third-party test frameworks.
- Keep tests hermetic: do not make real network calls in unit tests — use `net/http/httptest` or the `test-server/` for manual verification.
- Every new behavior in `data/`, `msg/`, `utils/`, or flag parsing should come with a test. Run `go test ./...` before considering a task done.

## Important Gotchas

- **Concurrency model**: workers write per-aggregate `RequestStats` to a buffered `respChan`; the main loop aggregates until `RespNum == Thread`. `Ctrl+C` (OS interrupt) triggers `cancel()` to end the test early — preserve this behavior when touching `main.go`.
- **Connection pooling** is deliberately sized from `-t` (thread count): `MaxIdleConns`, `MaxIdleConnsPerHost`, `MaxConnsPerHost` all equal the thread count. Don't hardcode these.
- **TLS/mTLS**: if any of `clientCert`/`clientKey`/`caCert` is set, all three are required (`genHttpClient` errors otherwise). `skipVerify` must stay opt-in.
- **Status code handling** in `doReq`: only 200/201 count as success; 4xx/5xx become typed errors. Keep stats semantics consistent if you extend it.
- Output flags interact: `-format json|csv` suppresses the incremental `raw` printing (`printLByFormat`), and `-resFile` writes the collected `Res` lines at the end. Test both when changing output logic.
- README is bilingual (Chinese + English) — update both language sections when user-facing behavior or CLI options change.

## Definition of Done

1. `gofmt -l .` reports nothing, `go vet ./...` is clean.
2. `go test ./...` passes.
3. README/`msg/` updated if user-facing behavior or strings changed.
4. `go build` succeeds with `CGO_ENABLED=0`.
