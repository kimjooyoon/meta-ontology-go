# Generate an output-error guard from Gooo

This experimental compiler adapter turns an explicit Gooo repair choice into a
Go source candidate. It does not execute the candidate or edit the input project.
It is an optional computes profile, not a new first-class policy syntax.

## The problem it closes

Before PR #849, a successful Gooo package execution could return exit code zero
even when writing its human-readable result failed. JSON output already handled
that failure. This profile can reuse that existing, explicitly selected handler
for the adjacent human-output branch.

The original fix was assistant-authored Go. This later capability must not
retroactively label that fix as autonomous language self-improvement.

## Declare the choice

Supply the exact SHA256 digest of the Go source bytes and of the existing
handler block, including its braces and whitespace. The placeholders below are
not executable digests.

~~~gooo
package goerrorguard
namespace goerrorguard
entity Source id "gooo://error-guard/source"
entity Candidate id "gooo://error-guard/candidate"
activity GuardWrite(Source) -> Candidate computes "go-error-guard:v1;function=writeSourcePackageResult;writer=stdout;diagnostic=stderr;source=sha256:SOURCE_HEX;handler=sha256:HANDLER_HEX"
~~~

The existing frontend lowers this program to semantic IR. The host requires the
activity's Source use and Candidate generation relations to match that IR.
Function, writer, diagnostic writer and both pins come from Gooo, not a guessed
natural-language instruction.

~~~sh
go run ./examples/meta-policy-compilation/go-error-guard \
  -program /caller/input/guard.gooo \
  -source /caller/input/original.go \
  > /caller/output/proposal.json
~~~

The adapter writes JSON to stdout only. Candidate bytes are in
`candidate_source`; `edit_start` and `edit_end` are zero-based source-byte offsets.
The caller controls any subsequent output, application or execution.

## Deliberately small support boundary

The target is a free, non-generic function with one int result and two distinct
io.Writer parameters. An adjacent branch must contain one existing
`if _, err := writer.Write(...); err != nil` guard with a diagnostic Fprintf and
one return value. The selected sibling contains one discarded fmt.Fprintf call
to the same writer. Import and parameter bindings are checked syntactically;
closures and unsupported shapes are not guessed through.

Only the selected call span is replaced. The handler is copied byte-for-byte
and the original call is evaluated once. A proposal is not a whole-program type
proof, semantic-equivalence proof, verified repair or permission to apply it.

`UNKNOWN` records stage, step, reason, unknown_class, next_operation and
blocked_by. Missing functions, unsupported shapes and ambiguity are distinct.
Contradictory source or handler pins are `REFUTED`. Changed candidate source
cannot silently repair the original pin or become a claimed fixed point.

## Native before/after case

The checked-in Go golden fixture is the actual original
`cmd/gooo/run_package_source.go` at a97a1025d50c5405302853155ae2f4e93436d0cf.
The native test freezes the unchanged repository regression file introduced by
#849 before generating a candidate. Both trials use the same oracle bytes.

GitHub Actions builds the original and generated source using caller-owned Go
build overlays. It runs only the four selected leaf paths in cmd/gooo:
human rejected writer, JSON rejected writer, human success, and JSON replay.
The required before pair is 3 pass / 1 fail; the candidate must produce
4 pass / 0 fail. These are acceptance conditions until actual native events
exist, not prefilled observed results.

A subsequent Go build and test process consumes the generated source. Native
event bytes, stderr, process exits and source/oracle/program/IR digests are
logged. Temp artifacts do not change runtime filesystem reads, and are not
persistent published release artifacts.

The Gooo program materialized by the tests is a caller-input fixture, not a
checked-in production language corpus entry or new completeness credit.
The oracle is separate from proposal generation but uses the same Go ecosystem,
not a diverse independent implementation. Adoption, external utility and
performance improvement remain UNKNOWN. No CI policy, fixed denominator,
local Go execution or automatic repository repair is added by this change.
