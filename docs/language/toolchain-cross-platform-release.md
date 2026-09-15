# Toolchain cross-platform release

## Boundary

The v1 claim is intentionally narrow: one exact source SHA produces native release
candidates on Linux amd64, Darwin amd64, and Windows amd64. It does not claim every
Go port and it does not publish a GitHub Release.

GitHub documents the fixed runner labels used by this corpus:
https://docs.github.com/en/actions/reference/runners/github-hosted-runners

Go documents build flags and environment variables here:
https://pkg.go.dev/cmd/go

## Meta operation

`assemble-exact-cross-platform-release` consumes three external platform receipts.
The receipts are facts; only the aggregate operation can grant readiness credit.

Each receipt binds:

- exact Git commit and Go 1.27.0
- native GOOS and GOARCH
- a clean VCS build with `-trimpath` and `CGO_ENABLED=0`
- two byte-equal binaries
- two byte-equal deterministic archives
- one native `gooo version --json` execution
- zero repository writes and zero mutation authorities

The aggregate emits one sorted `SHA256SUMS` file and three archives. Unknown top
decisions lower resolution and fail closed rather than becoming a fixed point.
