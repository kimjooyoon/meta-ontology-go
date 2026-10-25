# Go ecosystem observation and agent use

Gooo's source and PROV-O records describe why a semantic choice exists.
Performance observations describe what happened during one execution. The
two records can share an exact commit or receipt reference without giving a
timing measurement authority over language meaning.

## Read the signals already available

| Question | Existing source | Limit |
| --- | --- | --- |
| Who contributed a change? | GitHub commits and pull requests | Authorship is not a performance measure. |
| Which CI check took time? | GitHub Actions job and step timestamps; `ci-effort-observation` artifact | Queue time and job runtime are separate; one run does not prove a trend. |
| Which Go package/test consumed time? | Native `go test -json` output and existing command-origin artifacts | Package elapsed values are not an additive CI critical path. |
| Did an implementation change improve performance? | Pinned Go benchmarks with `-benchmem` and a same-workload comparison | A different toolchain, machine, input, or cache state breaks comparability. |
| What happened inside an agent call? | Optional ADK/OpenTelemetry trace | A trace is not a Gooo semantic decision or provenance proof. |

For a specific CI run, the GitHub CLI already exposes job intervals:

```sh
gh run view RUN_ID --repo kimjooyoon/meta-ontology-go \
  --json jobs --jq '[.jobs[] | {name,startedAt,completedAt,conclusion}]'
```

The [CI effort workflow](../.github/workflows/ci-effort-observation.yml)
publishes an `ci-effort-observation-<head-sha>` artifact, including its exact
source run binding, workflow window, required-check bottleneck and baseline
status. Use the observer's run ID when downloading that artifact. A pending or
`UNKNOWN` observation must not be presented as a measured improvement.

Go benchmarks and profiles remain available through the standard toolchain.
Choose a stable package and input, then retain the toolchain version and both
raw outputs with the compared commits:

```sh
go test ./internal/bidir -run '^$' -bench . -benchmem
go test ./internal/bidir -run '^$' -bench . -cpuprofile cpu.prof
go tool pprof cpu.prof
```

These commands are an operator recipe, not an additional required CI gate.
The existing CI effort receipt is the first place to look for CI cost; add a
new measurement only when it answers a question that receipt cannot.

Use the formatter shipped with the selected Go toolchain as well:
`"$(go env GOROOT)/bin/gofmt" -w .`. Automatic toolchain selection by `go`
does not replace a standalone `gofmt` on the shell's `PATH`; different
formatter versions can produce different source bytes in local and CI runs.

## Telemetry and ADK boundaries

[Go Telemetry](https://go.dev/doc/telemetry) collects usage and performance
information for Go-team-maintained tools and selected dependencies. Its local
mode can help inspect Go tooling; it is not a Gooo upload service. Gooo does
not enable telemetry upload on behalf of users.

Inspect the current Go setting with `go telemetry` and the local data location
with `go env GOTELEMETRYDIR`. Choosing `go telemetry on` contributes the approved
Go-tool counters to the Go team; it does not publish Gooo usage or benchmark data.

For runtime traces or metrics, [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
and OTLP provide a standard export path if an application needs one. Export
should be explicitly configured by the host application, with bounded
attributes and no source text or model prompts.

The optional [Google ADK Go bridge](../integrations/adk-go/README.md) exposes
the stable `gooo check --semantic --json` command and deterministic capability
discovery as separate function tools. It is a separate Go module, so the
compiler and users who do not use ADK do not load
its dependencies. The tools return Gooo's original JSON reports; the model's
choice to call them is distinct from the compiler's deterministic result. A
catalog match is not an implementation proof. The bridge creates a child
OpenTelemetry span for each compiler call when the host provides a tracer.
Agent hosts can inspect invocation timing without making it a semantic fact.

The [bridge CI](../.github/workflows/adk-go-bridge.yml) builds the Gooo CLI
from the submitted revision and calls it through the Go adapter with valid
source, a semantic error, and capability discovery. It runs when the adapter,
CLI, internal compiler code, or root Go module changes. This is command
compatibility evidence, not a model evaluation or runtime performance trend.
A manual run of the same workflow retains raw native Go benchmark output and
the Go toolchain version; normal PR runs do not add benchmark time.

The optional module's `BenchmarkCompiler` compares the same compiled Gooo CLI
and inputs through direct and bridge calls. Its valid-check, semantic-error,
and discovery workloads reuse the command compatibility fixtures and compare
the exact JSON and exit status on every call. The standard `ns/op` includes
compiler process startup; allocation values cover the parent process only.
Manual CI retains the actual runner image, architecture, compiler build settings,
and setup-go cache result with five raw samples per route. These fields help
separate environment changes from implementation changes when using
[benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat). A runner label
such as `ubuntu-latest` alone does not identify the image used in a run;
[GitHub documents the setup log as the source of image identity](https://github.com/actions/runner-images#what-image-version-is-used-in-my-build).

## Go module ownership in repository metrics

`go run ./scripts/line-metrics -json` includes Go files from nested modules in the
repository totals. Nested Go files, function indicators, and duplicate groups carry a
repository-relative `module_root` pointing to the nearest owning `go.mod`;
the root module is implicit to keep the report compact. Aggregate observations
retain structured source paths so ownership never depends on parsing display text.
A Go meta operation
resolves imports through that owning module rather than borrowing the root
module's dependencies. A source symlink crossing module roots remains in the
line totals, but its rewrite indicator is `NOT_APPLICABLE` with a module
boundary reason. The extractor independently rejects that write path.
If the owning `go.mod` requires a newer Go version than the fixed local
toolchain, the rewrite indicator is likewise `NOT_APPLICABLE`; the report
still counts the file. Gooo does not download a toolchain during metric scan.
The declaration splitter, density rewriter, assignment refactor, and repository extractor also
check the local toolchain when called directly. Recipe edits and creations
must stay in the source's module, and prospective imports resolve through that
module's dependencies. File or directory aliases may not hide a cross-module
write; aliases within one module remain supported. A module-boundary exclusion
remains visible in the plan and does not invalidate unrelated actionable indicators.

The boundary checks have both directions: a dependency declared only in the
nested module resolves, and a dependency declared only in the root module does
not resolve for the nested source. The measured lines, module identity, and
rewrite applicability are separate observations; none changes Gooo source
intent.

The module boundary determines whether a rewrite is eligible; it does not
implement an otherwise unsupported operation. In particular, duplicate groups
retain their source-fragment subject kind and the planner's existing operation
input checks. Package relocation uses a separate disposable-topology recipe;
that recipe's cross-module moves still need an explicit module contract.
[#1761](https://github.com/kimjooyoon/meta-ontology-go/issues/1761) tracks that
remaining path and exact-head observations of the hosted consumers.
