# Contributing to shipnote

Thanks for your interest. shipnote is pre-alpha; the surface area is still moving fast.

## Ground rules

1. **Determinism is non-negotiable.** Same inputs must produce byte-identical outputs. No `time.Now()`, no map iteration in renderers, no unsorted output.
2. **No outbound network calls** outside `internal/collect/github` and `internal/ai`. A lint rule enforces this.
3. **Apache-2.0 only.** Do not introduce dependencies under copyleft licenses (GPL, AGPL).
4. **Discuss before building.** For any change beyond a clear bug fix, open an issue first.

## Development

```bash
make build      # build the binary
make test       # run unit tests
make golden     # run golden-file determinism tests
make lint       # golangci-lint
make fmt        # gofumpt
```

Requires Go 1.22+.

## Pull requests

- One concern per PR.
- Add a `changelog:` block to the PR body (see the PR template).
- Add a label that maps to a category in [.shipnote.toml](./.shipnote.toml) once it exists.
- All commits should pass `go vet`, `golangci-lint run`, and `go test ./...`.
- Sign-off is not required, but the [DCO](https://developercertificate.org/) applies in spirit.

## Architecture decision records

Significant design changes go through an ADR in `docs/adr/`. Copy `docs/adr/0000-template.md`, give it the next number, open a PR.

## Reporting security issues

See [SECURITY.md](SECURITY.md). Do not open public issues for vulnerabilities.

## License

By contributing, you agree your contribution is licensed under [Apache-2.0](LICENSE).
