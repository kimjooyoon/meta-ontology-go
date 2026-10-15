# Source-control reader: first local observations

The [Gooo program](../../../examples/source-control-reader/README.md) interprets
retained IR facts with its own conditions. Its first native suite matched 15/15
new finite cases in 1.450 seconds including the local build. Model calls and
source-producer executions were zero.

The first code-generation request exposed an `else` placement error, then passed
after the source correction. The first pretty-printed case file exceeded the
32 KiB input bound and was rejected before execution. Compacting the same new
cases to 19,268 bytes allowed their first native invocation. These failures and
the original outputs remain in `first-cohorts.tar.gz`.

`observations.json` gives each original local command's scope and timing.
`archive-manifest.json` identifies the 33 byte-exact source/output members.
Only the Go executable location in published command metadata is shortened;
the original command-record digest is retained. The archive is 131,207 bytes.

The v81/v82 registrations appended two source cases, preserving the previous
118 cases and 59 historical denominator files. Registration proposals have
semantic admission `UNASSESSED`; current-source acceptance belongs to the new
original PR CI. Static truth uses retained guard nodes as its denominator;
the finite suite measures those 15 expected results.
