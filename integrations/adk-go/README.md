# Gooo tools for Google ADK Go

This optional module exposes `gooo check --semantic --json` and deterministic
`gooo discover --json` as ADK Go function tools. It does not add ADK to the
compiler's root module, change Gooo source, or affect deterministic checks
when no agent is present.

Build Gooo from the repository root:

```sh
go build -o /absolute/path/to/gooo ./cmd/gooo
```

Register the tool in an ADK Go agent:

```go
checkTool, err := adkcheck.New(adkcheck.Config{
    GoooBinary: "/absolute/path/to/gooo",
    SourceRoot: "/absolute/path/to/project",
})
if err != nil { return err }
discoverTool, err := adkcheck.NewDiscover(adkcheck.Config{
    GoooBinary: "/absolute/path/to/gooo",
    SourceRoot: "/absolute/path/to/project",
})
if err != nil { return err }
// Add both tools to llmagent.Config.Tools in the host application.
```

Import `github.com/kimjooyoon/meta-ontology-go/integrations/adk-go` as
`adkcheck` from a Go module that depends on this optional module. The tool
accepts a source-root-relative `.gooo` path. It returns `report_json` with the
original Gooo JSON bytes and `exit_code` from the command; source diagnostics
are returned even when Gooo exits nonzero. The agent cannot choose the binary
or source root. Symlinks that escape the configured root are rejected. Output is
limited to 1 MiB per stream (stdout and stderr each) and execution has a
15-second default timeout. Gooo receives
the relative path, so the configured absolute source root is not written into
the JSON report.

`gooo_discover_capability` accepts `path`, a `query` of at most 512 bytes, and
an optional `domain_contract` path. Both paths are checked against the same
source root. It returns the original `gooo discover --json` receipt and exit
code. The tool does not accept a generation artifact, executable cases, or a
toolchain override; it cannot start a native build. An `AVAILABLE` catalog
match describes a known capability, while the receipt keeps unverified
implementation or domain coverage `UNKNOWN`. Neither tool authorizes a
repository change.

For both tools, a valid JSON report with a nonzero exit code is a Gooo result,
not an ADK transport failure. A rejected path, invalid query, timeout,
oversized output, or malformed compiler JSON is returned as a tool error.

Boundary errors have a stable `gooo-adk/<stage>/<code>` string and an exported
`adkcheck.Failure` type (`Stage`, `Code`, `Unwrap`). The bridge does not copy a
source path, command output, or model prompt into that public error string.
ADK returns tool errors separately from successful tool result maps, so a
compiler diagnostic remains in `report_json` with its nonzero `exit_code`.

| Stage | Codes | Next action |
| --- | --- | --- |
| `configuration` | `invalid_binary`, `invalid_source_root`, `unavailable_root` | Host fixes the fixed executable or root. |
| `input` | `invalid_path`, `invalid_contract_path`, `invalid_query` | Agent corrects the bounded argument. |
| `execution` | `deadline`, `cancelled`, `start_failed`, `terminated` | Host checks cancellation, timeout, or executable availability. |
| `output` | `output_limit`, `invalid_json` | Host inspects the compiler/protocol boundary. |

Run `go test -run '^$' -bench 'Benchmark(Bridge|Direct)' -benchmem` in this
module for native Go timing/allocation observations. The fixed shell fixture
does not run the compiler or a model. The bridge and direct process benchmarks
include process startup; `BenchmarkBridgeRejectedQuery` measures validation
without starting a process. These observations do not establish compiler
speed or a CI critical-path improvement.

These tools observe the compiler's decision. ADK tool-call traces and timing are
execution observations, while the Gooo JSON report and PROV-O records retain
their own semantic meaning. No telemetry export is enabled by this module.

The optional bridge CI builds the CLI from the same commit and exercises both
tools without a model. Locally, build `gooo` and pass its absolute path as
`GOOO_ADK_TEST_BINARY` to `go test ./...` in this module to run the same
command compatibility checks.
