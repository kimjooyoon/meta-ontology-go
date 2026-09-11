# Candidate authorization bridge

This contract connects the read-only `NonExecutingImprovementCandidate` to the
existing `gooo/semantic-self-adoption-authorization/v1` authorization shape.
It does not authorize execution, repository mutation, promotion, or automatic
adoption.

The bridge has three explicit semantic transitions:

1. `RequestCandidateAuthorization` binds the candidate artifact, observation,
   policy, contract, subject, and scope digests into one request.
2. `DecideCandidateAuthorization` consumes an explicit decision input. A
   missing decision is `UNKNOWN`; a contradictory binding is `REFUTED`.
3. `ResolveCandidateAuthorization` emits a closed allow or deny receipt only
   after exact validation. Both outcomes keep execution and repository writes
   disabled.

The canonical denominator is fixed at nine cases: three `CLOSED`, three
`UNKNOWN`, and three `REFUTED`. Canonical fixtures are not live human
decisions. Live evidence is reported separately as `live_authorized=0` and
`live_state=UNKNOWN` until a workflow-dispatch decision is supplied.

GitHub actor and run metadata are recorded for workflow-dispatch decisions as
unsigned provenance metadata. They are not presented as cryptographic identity
or signature verification.
