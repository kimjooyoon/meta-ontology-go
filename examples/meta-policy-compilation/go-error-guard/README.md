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
Before generation, the native test uses Go's active-package file list to freeze
the current CLI Go files and its complete active test-source set. Both trials
use the same oracle bytes, including tests moved into generated files.
Runtime billing-package Gooo input digests are recorded and checked unchanged.

The overlay resolves the function named by the Gooo program in that active
package. It replaces only that declaration in its current owning file, not the
whole historical file. Package, signature and required import bindings must
match. Missing or duplicate declarations and changed context fail explicitly.
Existing extracted files and unrelated declarations are neither removed nor
reintroduced. The input files and active file membership are checked after
both trials. This preserves the repository-projection and nested-verifier
paths rather than skipping them.

GitHub Actions builds the original and generated declaration variants using
caller-owned Go build overlays. It runs only the four selected leaf paths in cmd/gooo:
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
The native case adds two bounded Go package-list observations around the two
focused test trials, not additional full-suite builds. The oracle digest now
identifies the active test-source digest map (ACTIVE_TEST_SOURCE_SET), not one
physical test file. This snapshot covers direct package Go files and the named
runtime Gooo inputs, not the entire dependency or toolchain closure.

The oracle is separate from proposal generation but uses the same Go ecosystem,
not a diverse independent implementation. Adoption, external utility and
performance improvement remain UNKNOWN. No CI policy, fixed denominator,
local Go execution or automatic repository repair is added by this change.

## V2: a local writer interface and a return-only handler

V1 is unchanged. Select V2 explicitly when a free function returns from its
guarded branch and then prints the human result before its final return:

~~~gooo
package goerrorguard
namespace goerrorguard
entity Source id "gooo://error-guard/source"
entity Candidate id "gooo://error-guard/candidate"
activity GuardWrite(Source) -> Candidate computes "go-error-guard:v2;function=runMetaPolicyGenerationProfile;writer=stdout;writer-type=interfaceWriter;source=sha256:SOURCE_HEX;handler=sha256:HANDLER_HEX"
~~~

The placeholders are not executable pins. V2 has exactly five fields:
function, writer, writer-type, source and handler. The writer type must be an
explicitly selected, non-generic, non-aliased local interface with exactly
Write([]byte) (int, error). Embedded or imported interfaces and changed method
sets are not inferred. The copied error handler must contain only one return.
The guarded branch, discarded Fprintf and final return must be consecutive at
the end of their block. Ambiguity and unsupported shapes remain UNKNOWN.

The actual public generation and policy-revision functions at accepted source
ad3af2b2f70ca2dfd3cb477b5cea7cce855c9032 are retained as one original Go fixture.
Both are proposal cases; this change's native behavioral case covers public
generation only. It must not be reported as policy-revision runtime coverage.

The native case adds one frozen test file through an external build overlay,
without writing a test into the repository. Its effective oracle digest covers
the complete active test-source map plus that exact additional file. The
runtime policy.gooo input is pinned and checked unchanged, separately from the
base snapshot's billing inputs. The criteria are five leaf paths: human and JSON
delivery to accepted and rejected writers, plus the existing exact-four-file
regression. The required original result is 4 pass / 1 fail; a generated candidate
must produce 5 pass / 0 fail. Counts are requirements until CI evidence exists.

Successful human and JSON bytes must remain exact. Rejected delivery must not
erase already-generated external files or change the input source. The native
case uses two additional focused Go test processes, not another full suite.
It follows declarations after repository projection rather than replacing the
whole historical file. A proposal and a passing temporary build are still not
persistent source adoption. Recover candidate bytes from CI, then obtain the
separate public-profile owner's native acceptance before adoption. No source
write permission, CI skip, promotion authority or performance claim is added.


## Opt-in canonical Go pipeline

A separate pipeline entry point connects the guard and renderer through the
existing Gooo semantic graph. The default adapter and single-activity API still
require two entities and one activity; they never guess this larger contract.

~~~gooo
package goerrorguard
namespace goerrorguard
entity Source id "gooo://error-guard/source"
entity RawCandidate id "gooo://error-guard/raw-candidate"
entity Candidate id "gooo://error-guard/candidate"
activity GuardWrite(Source) -> RawCandidate computes "go-error-guard:v2;function=runMetaPolicyGenerationProfile;writer=stdout;writer-type=interfaceWriter;source=sha256:SOURCE_HEX;handler=sha256:HANDLER_HEX"
activity Canonicalize(RawCandidate) -> Candidate computes "go-source-format:v1;toolchain=go1.27.0"
~~~

SOURCE_HEX and HANDLER_HEX are placeholders, not executable pins. The guard may
select the existing v1 or v2 profile without changing that profile's rules.

~~~sh
go run ./examples/meta-policy-compilation/go-error-guard \
  -pipeline \
  -program /caller/input/canonical-guard.gooo \
  -source /caller/input/original.go \
  > /caller/output/pipeline-proposal.json
~~~

The guard's raw candidate stays under the guard field. The top-level
candidate_source is the canonical result of the explicitly declared renderer.
Both activity IDs and the full input program/semantic digests are retained.
The formatter version must exactly match the declared Go runtime version;
this is a toolchain selection boundary, not a binary signature or provenance
proof for an arbitrary host.

Rendering requires the original source to already be canonical. It rejects
formatting changes outside the original guard edit span instead of normalizing
unrelated source. Its three native go/format calls check the original, render
the raw candidate, and check the formatted candidate's fixed point. A failed
guard does not invoke the renderer; it retains the direct failure and the
renderer dependency-blocked UNKNOWN frontier separately.

A rendered result remains a PROPOSED candidate with UNKNOWN independent
admission and zero mutation/promotion authority. Formatter fixed point is not
semantic equivalence, a repair acceptance, an execution receipt, or an
automatic repository change.

The native pipeline case uses the same five public-generation leaves and
runtime-input protections as the raw v2 case, but its after variant consumes
the pipeline's canonical output. Before4PASS/1FAIL and after5PASS/0FAIL remain
requirements until observed in CI. It adds two focused CLI trials, not another
full-suite invocation. The adapter has separate default/explicit-opt-in tests.
No local Go invocation or CI/ownership-policy change accompanies this feature.


## V3: stop a declared human-output branch at its first delivery failure

This explicit profile addresses the actual normal-generation reporter
reportGenerateSuccess, rather than relabeling a synthetic formatting example
as public adoption. Its original Go source is frozen from
cmd/gooo/generate_pipeline_part03.go, last changed at
f8005181f66921d6668eb8465325b21308a0eef8. That reporter has three conditional
human-output sites and a separate JSON-handler return.

~~~gooo
package goerrorguard
namespace goerrorguard
entity Source id "gooo://error-guard/source"
entity RawCandidate id "gooo://error-guard/raw-candidate"
entity Candidate id "gooo://error-guard/candidate"
activity GuardHumanOutput(Source) -> RawCandidate computes "go-error-guard:v3;function=reportGenerateSuccess;writer=stdout;mode=jsonMode;handler-call=writeJSONReport;writes=3;source=sha256:SOURCE_HEX;handler=sha256:HANDLER_HEX"
activity Canonicalize(RawCandidate) -> Candidate computes "go-source-format:v1;toolchain=go1.27.0"
~~~

Use the existing explicit -pipeline adapter. SOURCE_HEX and HANDLER_HEX remain
non-executable placeholders. The seven v3 fields select the function, io.Writer
parameter, bool mode parameter, existing JSON handler call, exact positive
write-site count, source digest and handler-block digest. Nothing is selected
from a natural-language title. V1 and v2 matching are unchanged.

The first statement must be if !mode with no initializer or else. Its body
contains discarded fmt.Fprintf calls to the selected writer, optionally inside
initializer-free/else-free conditional blocks, followed by the same one-value
return as the function's final return. Conditions cannot invoke functions,
receive from channels or reference the output writer. The writer cannot
escape through another formatting argument. Loops, concurrency, closures as
statements, local binding changes and unsupported writer types are not guessed
through. The explicitly named JSON function must use the selected writer and a
return-only error guard; a shadowed local handler is unsupported.

The compiler copies that pinned handler around each declared write, preserving
evaluation order and existing bytes between writes. edit_start/edit_end enclose
the selected call sites; bytes outside that interval are unchanged.
guarded_calls is the generated number of guarded sites, not a test count or
quality score. A mismatched declared count or digest is REFUTED. Missing or
unsupported selection retains the existing six-field UNKNOWN cause.

The independent frozen native oracle has exactly eight leaves: four actual
public generate human/JSON accepted/rejected-output cases, plus four direct
reporter cases covering second-write failure, third-write failure, all messages
delivered and absent discovery. The public cases retain exact generated file
bytes, parse the generated Go, and preserve their Gooo runtime input. Conditional
cases require the exact delivered prefix and no later writes after failure.
A successful baseline call freezes each public case's expected artifacts before
its delivery trial; this does not add Go build/test invocations.

Required native results are before5PASS/3FAIL and canonical after8PASS/0FAIL.
These are requirements until CI observes them. The native test uses two focused
Go test processes with the same external overlay oracle, not another full suite.
It follows the actual declaration after projection and retains the base view.

This prepares compiler capability only. No public production reporter is changed
here. Candidate publication, the separate declared public-profile owner, full
native acceptance and subsequent accepted execution are still needed for
adoption. Neither the Gooo declaration nor canonical formatting can accept its
own repair, lower its oracle, authorize repository writes or claim utility or
performance improvement.
