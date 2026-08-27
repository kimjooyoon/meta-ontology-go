# Meta-circular boundary

This is a finite, read-only experiment for the proposition
`self-description != authorization != execution`. `main.gooo` declares the
meta-operation vocabulary and four executable `computes` values, but those
values contain authorization requests only. They do not contain an issuer,
grant, handle, or execution permission.

The Go contract fixes only four case identities and their predicate denominator:

1. description request without an external grant;
2. explicit external READ_ONLY grant;
3. forged issuer/handle grant;
4. external grant outside the READ_ONLY scope.

Producer and consumer parse/lower the `.gooo` computations independently. The
consumer shares only the versioned wire model and general Gooo parser/lowerer;
its case derivation, expected predicates, receipt reconstruction, summaries,
indicators, and source graph replay are separate. CI measures the boundary with
`go list -deps`: forbidden producer dependencies observed `0`, allowed maximum
`0`, independence contract `1/1`. A shared-wrong-case regression is rejected.

## Evidence contract

The external grant producer is `external-authority-fixture`. CI creates a raw
grant artifact after observing the source semantic digest and retains the raw
artifact digest. The evaluator never synthesizes a grant or handle. Missing
grant evidence is `OPEN / LOWER_RESOLUTION` (or an explicit policy deny).

The lowered graph must reconstruct the typed ordered path
`DescribeMetaOperation -> GrantReadOnlyMetaCapability -> ExecuteMetaOperation`.
Only an external grant whose subject, issuer, operation, scope, and handle pass
policy can authorize the defined `meta-circular-boundary.evaluate` operation.
That operation emits a canonical result and an execution artifact with path,
operation ID, grant digest, input digest, output digest, and artifact seal.
Blocked cases have no execution artifact. The execution claim is bound to the
actual output digest, not to an `Execution=ALLOWED` string.

CI observes tracked and untracked workspace state before and after the run,
the output outside the repository, and `workflow-contents-read-only` permission
evidence. Unknown permission or output placement is retained as
`OPEN / LOWER_RESOLUTION` at the exact effect coordinate. Receipt self-sealing
(`receipt_self_seal_valid`) is separate from `replay_matches`: replay is based
on two actual runs, `cmp`, and separate A/B receipt and output-digest arrays.

The producer emits no independent-judge result. The consumer emits a separate
judge receipt after comparison, and CI combines producer and consumer receipts
with `0` mismatches. Every case has description, authorization, and execution
propositions. Append-only transitions include proposition digests, external
grant evidence digests, execution output evidence digests, and dependencies.
Unknown evidence remains `OPEN`; contradictory capability evidence is
`REFUTED` only on the affected authorization claim.

## Causality and limits

CI records raw source/grant changes, semantic digests, grant digests, graph
digests, output digests, and claim transitions for five interventions:

| Intervention | Expected result |
| --- | --- |
| request-only | semantic request change lowers authorization/execution and claims |
| grant-change | grant scope changes while source semantic digest stays fixed; authorization is denied |
| description-only-forgery | authority-looking self-description is observed as escalated and blocked |
| comment-only | raw source changes while semantic, grant, output, and claim results are preserved |
| graph-connection | typed relation change lowers the graph/case result |

The experiment is inspired by the staged evaluator material in MIT's
[Meta-Circular Evaluator](https://groups.csail.mit.edu/mac/classes/6.001/FT98/lectures/lec16/eval.pdf)
and by E's object-capability security notes, where authority is conveyed by
explicit object references rather than ambient names:
[Capability Security](https://erights.org/elang/kernel/auditors/index.html).
Those sources motivate the boundary but do not prove cryptographic
unforgeability, general capability confinement, arbitrary Gooo execution, or
self-hosting. The falsifiers are executable: source-only grants, forged or
out-of-scope grants, fake ALLOWED strings, missing artifacts, changed graph
edges, and a consumer accepting a shared wrong case fact must fail or lower
resolution.
