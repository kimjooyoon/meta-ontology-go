# External capability authorization

This use case asks a narrower question than capability execution: may this
specific CI request use the already observed external meta-operation? The Gooo
policy declares ten evidence-producing activities. A small Go checker consumes
the compiled policy tree, the request envelope, and the `10/10` execution
report. It starts from deny and never grants repository mutation or promotion.

## CAB-10 fixed denominator

| # | Class | Choice | Meta-operation | Exact target |
| --- | --- | --- | --- | --- |
| 1 | driver | FOUNDATION | bind execution report | sealed `10/10` report |
| 2 | outcome | COHERENCE | bind requested operation | pinned operation |
| 3 | driver | FOUNDATION | bind subject | exact CI head SHA |
| 4 | driver | FOUNDATION | bind issuer | pinned workflow issuer |
| 5 | outcome | COHERENCE | bind scope | exact three-capability scope |
| 6 | driver | FOUNDATION | bind policy foundation | immutable merged-CI locator |
| 7 | guardrail | REGRESSION | enforce default deny | `DENY` |
| 8 | outcome | COHERENCE | bind invocation | exact run and attempt |
| 9 | guardrail | REGRESSION | bind nonce | deterministic envelope nonce |
| 10 | guardrail | REGRESSION | enforce effect ceiling | all authority and writes `0` |

The bootstrap CI has local evidence for nine obligations but no prior immutable
foundation for the new policy itself. Its honest result is therefore
`FAIL_CLOSED / UNKNOWN / 9/10`. The unknown ledger identifies
`AUTHORIZE/policy-foundation`; nine claims are `DISCHARGED` and one remains
`OPEN`. A local evidence reader sees `9/9 / EXACT`, while an authorization
reader sees `9/10 / UNKNOWN`.

The versioned conformance suite has 20 cases: one authorized shadow, seven
unknown-evidence cases, and twelve known denials. `AUTHORIZED_SHADOW` still has
`NO_EFFECT`; cryptographic delegation, cross-run replay storage, repository
mutation, and production enforcement are explicit non-claims.
