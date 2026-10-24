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
| `input` | `invalid_path`, `invalid_contract_path`, `invalid_query` | Agent checks the argument; host checks root availability if it changed. |
| `execution` | `deadline`, `cancelled`, `start_failed`, `terminated` | Host checks cancellation, timeout, or executable availability. |
| `output` | `output_limit`, `invalid_json` | Host inspects the compiler/protocol boundary. |

Run `go test -run '^$' -bench 'Benchmark(Bridge|Direct)' -benchmem` in this
module for native Go timing/allocation observations. The fixed shell fixture
does not run the compiler or a model. The bridge and direct process benchmarks
include process startup; `BenchmarkBridgeRejectedQuery` measures validation
without starting a process. These observations do not establish compiler
speed or a CI critical-path improvement.
The timeout cancels the direct Gooo process; this bridge does not prove cleanup
of processes that Gooo itself may spawn.

These tools observe the compiler's decision. Each compiler process call creates
an OpenTelemetry span named `gooo.check` or `gooo.discover` under the host's
active ADK trace. It records only the exit code or a stable boundary failure
stage and code. It does not record source paths, query text, JSON reports, or
prompts. The host configures its own tracer provider and exporter; this module
does not enable export. A no-op provider keeps standalone use deterministic.
Trace timing is an execution observation, while the Gooo JSON report and
PROV-O records retain their own semantic meaning.

The optional bridge CI builds the CLI from the same commit and exercises both
tools without a model. Locally, build `gooo` and pass its absolute path as
`GOOO_ADK_TEST_BINARY` to `go test ./...` in this module to run the same
command compatibility checks.

The bridge workflow can also be started manually to retain raw Go benchmark
output as an artifact bound to the commit SHA and toolchain version. Normal PR
checks do not run benchmarks, because their wall time is not a stable gate.

To observe the real compiler, set `GOOO_ADK_TEST_BINARY` to an absolute path
to a binary built from the same revision, then run:

```sh
go test -run '^$' -bench '^BenchmarkCompiler/' -benchmem -benchtime=200ms -count=5
```

The native Go benchmark runs valid source checking, a semantic error, and
capability discovery through both the direct command and the bridge. Every
call must return identical JSON and exit status for that workload. `ns/op`
includes process startup and compiler execution. `B/op` and `allocs/op` count
the parent process only; they do not measure compiler RSS or child allocations.
The bridge route uses the host's default no-op tracer and no model. A configured
exporter or a full ADK agent loop needs its own observation.

Manual CI runs include these workloads and retain the compiler build information,
runner OS/architecture, image version, and setup-go cache-hit result alongside
the raw samples. The cache field describes setup-go restoration, not OS file-cache
warmth. Compare matching workloads and environments with Go's `benchstat`;
measurements from different runner images are separate cohorts. Fixtures are tiny
declaration programs, so the results do not establish large-project throughput.
