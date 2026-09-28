# Candidate eligibility bridge

This contract evaluates a read-only `NonExecutingImprovementCandidate` from
its exact report, policy, source observation, execution input, contract, and
artifact digests. The system derives eligibility only when every binding is
complete and consistent. Missing evidence stays `UNKNOWN`; contradictory
evidence is `REFUTED`.

The three semantic transitions are:

1. `RequestCandidateAuthorization` binds the candidate artifact, observation,
   policy, contract, subject, and scope into one request.
2. `DeriveCandidateAuthorization` computes eligibility from that exact system
   evidence. It takes no actor decision and has no manual decision input.
3. `ResolveCandidateAuthorization` emits the deterministic system receipt and
   an independent verifier checks its bindings.

The canonical denominator remains nine cases: three `CLOSED`, three `UNKNOWN`,
and three `REFUTED`. The report records system-derived decisions, deterministic
replay, and zero fallback acceptance. Candidate eligibility does not authorize
execution, tests, repository writes, promotion, or adoption.
