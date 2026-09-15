# Counterexample-derived Gooo execution

This example composes the accepted counterexample proposal with
`ObservePolicyDecisionRevision(PolicySource, RevisionRequest) -> RevisionObservation`.
The existing Gooo contract is parsed, lowered and checked against its native ABI.
No new registry, evaluator rule or adoption permission is introduced.

```sh
go run ./examples/meta-policy-compilation/counterexample-execution \
  -policy /caller/policy.gooo -input /caller/input.json \
  -operation internal/meta/policycompilation/revision-operation.gooo
```

The input object contains `counterexample` (the existing source/result-bound
counterexample) and `cases` (existing baseline/candidate case pairs). It contains
no caller-selected condition/from/to fields: those come from the counterexample.
At least one baseline must equal its complete original case, and each pair must
retain the same validator expectation. Candidate evidence must be supplied
separately; this adapter does not rewrite old evidence digests to fit new code.

The stdout report preserves the proposal, exact `revision_request_json` bytes
and nested Gooo operation/execution receipt. The request string and nested
observation can be passed to the existing independent receipt consumer.
That consumer checks consistency, not independent execution or adoption.
Empty case coverage retains its six-field UNKNOWN cause. Failed native work
retains both its earlier proposal and incomplete execution evidence.

Native CI exercises the public CLI, independent receipt consumer and a fresh
process consuming the emitted candidate judge. This is a synthetic, bounded
next-use witness, not autonomous persistent policy adoption or external utility.
Its three input pairs yield observed counters, not a language completion score.
Proposal-stage `NOT_OBSERVED` stays historical; later execution has its own
receipt. Input files remain immutable; generated execution uses temporary space.
No repository writes, promotion, runtime-speed improvement or cache proof is
claimed. The original four-artifact and two-artifact profiles remain unchanged.
