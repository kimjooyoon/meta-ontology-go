# Capability feedback investment aggregator

This command reads a JSON array of capability discovery feedback and emits grouped investment candidates.

## Input contract

Each feedback record must include:

- source digest;
- declaration digest;
- capability receipt digest;
- evidence digest;
- suggested question;
- outcome: useful, not_useful, or unresolved;
- non-executing and non-authorizing set to true.

Records with missing provenance, unsupported outcomes, trailing JSON, or execution/authorization flags are rejected.

## Output interpretation

Candidates are grouped by source digest, declaration digest, and suggested question. The output retains distinct evidence digests and counts each outcome.

- COLLECT_MORE_OBSERVATIONS means the evidence is insufficient for an investment review;
- REVIEW_INVESTMENT means at least two useful observations exist with no unresolved observation for that candidate.

REVIEW_INVESTMENT is only a bounded prioritization signal. It does not claim domain completeness, authorize a capability, or automatically change the language. A human or a later policy stage must inspect the evidence and record the investment decision and its falsifying observation.

The output is designed to feed the domain completeness receipt while keeping provenance and uncertainty visible.