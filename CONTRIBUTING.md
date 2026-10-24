# Contributing

Thanks for contributing to Desktop Notify.

## Development requirements

- Go 1.27 or newer
- GNU Make or a compatible `make`
- A desktop notification session for manual smoke tests

The project intentionally has no runtime Go dependencies and does not require
Docker or a database.

## Before opening a pull request

Run:

```bash
make fmt
make check
go test -race ./...
```

Changes should keep the public SDK small, avoid logging notification content,
and include tests for new behavior. Platform-specific behavior should be
isolated in the corresponding build-tagged backend and documented in
`README.md`.

## Commit and release conventions

Use clear, imperative commit messages. Releases are created from tags matching
`v*`; do not publish a release manually for ordinary branch pushes.
