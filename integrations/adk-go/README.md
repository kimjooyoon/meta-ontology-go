# Gooo check for Google ADK Go

This optional module exposes the existing `gooo check --semantic --json` command
as an ADK Go function tool. It does not add ADK to the compiler's root module,
change Gooo source, or affect deterministic checks when no agent is present.

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
// Add checkTool to llmagent.Config.Tools in the host application.
```

Import `github.com/kimjooyoon/meta-ontology-go/integrations/adk-go` as
`adkcheck` from a Go module that depends on this optional module. The tool
accepts a source-root-relative `.gooo` path. It returns `report_json` with the
original Gooo JSON bytes and `exit_code` from the command; source diagnostics
are returned even when Gooo exits nonzero. The agent cannot choose the binary
or source root. Symlinks that escape the configured root are rejected. Output is
limited to 1 MiB and execution has a 15-second default timeout. Gooo receives
the relative path, so the configured absolute source root is not written into
the JSON report.

This tool observes the compiler's decision. ADK tool-call traces and timing are
execution observations, while the Gooo JSON report and PROV-O records retain
their own semantic meaning. No telemetry export is enabled by this module.
