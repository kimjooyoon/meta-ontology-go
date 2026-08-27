# Meta-circular boundary experiment

Status: executable read-only experiment on PR #567. This is a boundary
experiment, not a self-hosting evaluator and not a general capability-security
claim.

## Proposition

Gooo can describe a meta-operation without thereby receiving permission to
execute it. The implementation keeps three propositions separate:

1. `self-description`: `main.gooo` is parsed, lowered, normalized, and bound
   to the source semantic digest. Its four executable `computes` values are
   authorization requests only.
2. `authorization`: an independently produced raw external-grant artifact is
   parsed by both sides. The source never creates its own grant, issuer, or
   handle. A grant is accepted only when its issuer, subject semantic digest,
   operation, scope, and handle satisfy the external policy.
3. `execution`: only the explicit external READ_ONLY grant and the lowered
   `Describe -> Grant -> Execute` typed path permit the defined
   `meta-circular-boundary.evaluate` operation. The operation emits a
   canonical value and an execution artifact containing its path, operation
   ID, input digest, output digest, and artifact seal.

The Go contract fixes only the four case identities and therefore the
predicate denominator `4`. Observed request values, grants, effect snapshots,
replay evidence, receipts, and claim states are inputs or recomputations.

## Evidence boundary

The external grant producer is `external-authority-fixture`; CI records the
raw artifact SHA-256 in `source.grant_artifact_digest`, and each grant receives
its own digest after parsing. The effect producer is
`ci-workspace-observer`: CI hashes tracked and untracked repository state
before and after, records the output outside the repository, and supplies
`workflow-contents-read-only` permission evidence. A missing or unknown
permission observation is `OPEN / LOWER_RESOLUTION` at
`OBSERVE_EFFECT:resolve-workspace-permission-and-output`.

Replay is evidence from two separately executed witness runs. The replay
command compares both receipt-digest arrays and execution-output-digest arrays;
`replay_matches=4` is not a count of hashing one receipt again. Receipt
self-sealing is reported separately as `receipt_self_seal_valid=4`.

The producer emits no independent-judge result. The consumer package has its
own copied wire model and independently replays source parsing, lowering,
case derivation, expected predicates, receipt construction, summary, and
indicator recomputation. It emits a separate judge receipt. CI combines that
receipt with the producer report and checks `0` mismatches. `go list -deps`
must observe forbidden producer dependencies `0`, with allowed maximum `0`
and independence contract `1/1`.

## Claims and transitions

Each case has distinct description, authorization, and execution propositions.
The append-only transitions carry proposition digests and evidence digests.
Authorization evidence is the external grant digest; an allowed execution
transition is additionally bound to the execution output digest. The execution
claim is `UNKNOWN` when authorization or execution is unobserved; contradictory
capability evidence is `REFUTED` only for the affected authorization claim.

Unknown source or case data therefore remains `OPEN / LOWER_RESOLUTION` with
an exact `stage/step/reason` coordinate. It is not converted to a successful
case by an expected-decision field. `ExpectedDecision` is validator data only;
the producer decision is derived from the observed case and evidence.

## Causality interventions

CI runs five independent interventions and records raw, semantic, grant,
output, graph, and claim digests:

| Intervention | Expected observation |
| --- | --- |
| request-only | semantic request change changes authorization/execution and claims |
| grant-change | external grant scope changes while source semantic digest stays fixed; authorization is denied |
| description-only-forgery | an authority-looking self-description is observed as escalated and blocked |
| comment-only | raw source changes while semantic, grant, output, and claim results are preserved |
| graph-connection | a typed relation change invalidates the lowered path and lowers the result |

The falsifiers are concrete: a source-only `GRANT` or handle must not pass; a
valid-looking `Execution=ALLOWED` string without an artifact must not pass; a
missing output artifact in a blocked case must be observed; changing the
graph relation must lower the case; changing only a comment must not change
semantic outputs; and a consumer report that shares a wrong case fact must be
rejected.

## Principles and limits from prior work

The evaluator side follows the meta-circular lesson that a language processor
can be expressed in the language it processes, while this experiment adds an
explicit evidence and authority boundary. The MIT Press presentation of
*Structure and Interpretation of Computer Programs* places evaluator and
compiler construction in its language-processor material:
<https://mitpress.mit.edu/9780262367622/structure-and-interpretation-of-computer-programs/>.

The security side follows the object-capability principle that authority is
carried by an unforgeable reference or object access, not by a name or a
description. The E-language material describes capability security and the
auditor model, including confinement claims:
<https://www.erights.org/elib/capability/overview.html> and
<https://erights.org/elang/kernel/auditors/index.html>.

Those sources motivate the separation; they do not prove this repository's
fixture is cryptographically unforgeable, that the runner is a confined
object-capability machine, or that arbitrary Gooo can execute. The experiment
claims only the finite, digest-bound observations retained by CI.
