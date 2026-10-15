# Complete race scope with the compiler package scheduled first

This follows [PR #1571](https://github.com/kimjooyoon/meta-ontology-go/pull/1571)
at `4dbdbc664418c35869f23501a2677df47bf370c2`. It retains that complete
language change: direct and grammar-derived body origins, source-owned readers,
origin query starters, registration roles and source-scoped type readers.

## Original observation

The original race job `116004690167`, run `38608271678`, attempt 1, was
cancelled after its unchanged 20-minute job limit. Its digest-bound
[raw log](original-job.log.gz), [job metadata](original-job.json),
[annotations](original-annotations.json) and
[event reconstruction](original-event-observation.json) are retained.

| Observed event | Count |
| --- | ---: |
| Packages started | 431 |
| Packages passed | 243 |
| Packages with no tests skipped | 187 |
| Tests and subtests passed | 8,448 |
| Tests and subtests skipped | 17 |
| Test failure events | 0 |
| Package terminal events still missing at cancellation | 1 |

The extractor package started at `17:54:40.675101246Z`; the job started at
`17:42:57Z`. Its last test event was at `18:03:04.155662412Z`. There were no
unfinished test events in the captured stream, but its package PASS event and
successful command exit were absent. The cancelled job cannot authorize a merge. All [39 original workflows](predecessor-terminal-originals.json.gz)
finished before PR #1571 was closed: 38 succeeded and the core run was cancelled.
These event counts describe this execution; nested test counts do not measure
independent language behaviors. The log does not establish a deadlock.

## Scheduling change

The same job now invokes:

```sh
go test -race -json ./internal/meta/repositoryprojection/extractor ./...
```

The compiler package receives the first explicit package argument. The complete
`./...` selector, race instrumentation, JSON events, job name, 20-minute budget,
checkout identity and fresh-build setting remain in the command contract.
The workload is selected once by Go's package loader.

The installed Go 1.27.2 package loader appends packages in pattern order and uses
`seenPkg` to omit a package already loaded by an earlier pattern. Its test runner
builds actions in that order and inserts ordered test barriers. This explains the
scheduling hypothesis; dependency builds and runner contention can still affect
start times. Whole-CI improvement requires this successor's own original run.

## New bounded observations

- [scope57213](scope57213/manifest.json) records two new command-contract tests:
  full selection and race identity remain required; narrowed or masked commands
  are rejected. Only these new tests ran locally.
- [selection57517](selection57517/ended.json) enumerated actual packages and test
  variants with both argument lists under the same candidate overlay. Both lists
  contained 919 unique entries, with zero omitted, added or repeated entries.
  The extractor moved from position 312 to position 0. This enumeration executed
  no tests and supplies no performance result. Its changed source bytes are retained in
  [scope57213/source.tar.gz](scope57213/source.tar.gz).
- [static57791](static57791/ended.json) requested a `SplitSeq` modernization in
  the new test. Its original diff and source are retained.
- [scope58201](scope58201/manifest.json) adds a repeated-selector counterexample
  and the modernized test source. Both new tests passed once.
- [static58331](static58331/ended.json) records zero diff and zero diagnostics
  across all 80 carried Go files. It executed no tests. Completed historical
  source, test and model cohorts were not locally rerun.

The successor keeps the entire predecessor feature. Original workflows, current
canonical checks, source-bound proof and native readiness must all be inspected
before acceptance. Publication and the prepared `0.6.35-dev` release remain
unproven until their own source revisions pass.
