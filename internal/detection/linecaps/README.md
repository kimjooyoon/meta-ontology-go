# Line-cap detector

`internal/detection/linecaps` is an independent re-check of the repository's
Go size policy and an optional refactorability signal. It does not import
`internal/verify` or change the repository verification entry points.

## API

```go
limits := linecaps.DefaultLimits()
report, err := linecaps.Analyze(root, paths, limits)
if err != nil {
	return err // invalid input or repository walk failure
}
if err := report.Err(); err != nil {
	fmt.Fprintln(os.Stderr, err)
}
```

An empty `paths` slice discovers `.go` files below `root`, in lexical order,
excluding `.git` and `vendor`. `AnalyzeSource` is available when the caller
already has source bytes. `Limits` are inclusive: 300 file lines and 75
function lines pass; the 301st or 76th line produces a finding. Named functions,
methods, and function literals are measured from `func` through their closing
brace.

Paths supplied to `Analyze` are repository-relative (or absolute paths inside
`root`) and are normalized and deduplicated without changing the caller's
slice. Paths that escape `root` fail closed. Read and parse failures remain
findings so an incomplete re-check cannot silently pass.

## Output

`Report.Text` is stable human output:

```text
linecaps: violations=1
internal/example.go:12-87: function-lines Compute: got 76, limit 75
```

Refactorability findings are emitted with `refactor-return` and
`refactor-assign-return` rules:

```text
linecaps: violations=2
internal/example.go:10-12: refactor-return normalize: single return ReturnExpr (actual=1)
internal/example.go:20-23: refactor-assign-return makeValue: assignment then return value (actual=2)
```

`Report.JSON` is the machine-readable output. It contains a sorted `findings`
array with `path`, `rule`, `actual`, `limit`, and optional source range/name or
error detail fields.

### Layout/line metrics API

`AnalyzeLineMetrics` scans a directory tree and returns layout counts and line
metrics for `.go` and `.gooo` files:

```go
stats, err := linecaps.AnalyzeLineMetrics(".")
if err != nil {
	panic(err)
}
fmt.Println(stats.Total())
payload, _ := stats.JSON()
```

The directory rows include immediate and recursive folder/file counts, plus
recursive go/gooo file and line totals.

```text
line metrics: files=6 dirs=2 go_lines=7 gooo_lines=5
.: direct_folders=2 direct_files=0 folders=2 files=6 go_files=3 gooo_files=3 go_lines=7 gooo_lines=5
sub: direct_folders=1 direct_files=0 folders=1 files=4 go_files=2 gooo_files=2 go_lines=4 gooo_lines=2
sub/nested: direct_folders=0 direct_files=2 folders=0 files=2 go_files=1 gooo_files=1 go_lines=1 gooo_lines=1
```

You can also print this report directly from the repository root with:

```sh
go run ./scripts/line-metrics
go run ./scripts/line-metrics -json
```
