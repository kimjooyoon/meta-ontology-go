# Toolchain cross-platform release

## Boundary

One exact source SHA produces native release candidates on Linux amd64,
Darwin amd64, Darwin arm64 and Windows amd64. The checked-in corpus enumerates
these four targets. Publication follows the separate release contract.

GitHub documents the fixed runner labels used by this corpus:
https://docs.github.com/en/actions/reference/runners/github-hosted-runners

Go documents build flags and environment variables here:
https://pkg.go.dev/cmd/go

## Meta operation

`assemble-exact-cross-platform-release` consumes four external platform receipts.
The receipts are facts; only the aggregate operation can grant readiness credit.

Each receipt binds:

- exact Git commit and Go 1.27.1
- native GOOS and GOARCH
- a clean VCS build with `-trimpath` and `CGO_ENABLED=0`
- two byte-equal binaries
- two byte-equal deterministic archives
- native version, language-example, package, construction and saved-replay executions
- zero repository writes and zero mutation authorities

The aggregate emits one sorted `SHA256SUMS` file and four archives. Unknown top
decisions lower resolution and fail closed rather than becoming a fixed point.

The package construction profile checks a two-package imported helper at attempt
budgets five and six, then replays the saved six-attempt history. It preserves
partial 1/4 and complete 4/4 results, original package identities and caller rows,
local rejections, native faults and exact large integers. Native case counts are
reported separately from the structural release corpus.
