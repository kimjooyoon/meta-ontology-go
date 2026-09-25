# Source-backed authority shadow use case

This use case exercises the source-backed authority evaluator without granting
language-readiness credit or changing a repository gate.

## Pinned source

- Repository: `cosmos72/gomacro`
- Commit: `cf0d4bf32da393dbda97e3572f216731013ffa55`
- Source: `README.md#L1`
- Span bytes: `77`
- Span digest: `sha256:29362aa311de0f24c66f41cc65a8b6ffd996baf37e048b5a72db63172aae5bf2`
- Authority reference: `gomacro-readme-title-authority`

The pinned sentence is evidence only for gomacro's own stated description. It
does not establish a Gooo semantic claim or a novelty comparison.

## Deterministic observation

- Input facts: `2 = ACCEPTED 1 + CANDIDATE 1`
- Evaluator denominator: `1` accepted fact
- Backed facts: `1`
- Coverage: `10000 bps`
- Candidate leakage into denominator: `0`
- Replay digest mismatches: `0`
- Repository writes: `0`
- Gate effect: `NO_EFFECT`
- Promotion credit: `0 bps`

The CI adapter injects the exact PR head into the fixture, evaluates twice,
requires byte-identical receipts, and writes the artifact outside the repository.
UNKNOWN remains `UNKNOWN / INVARIANT_ONLY / BLOCK`; SHADOW mode never converts it
to a fixed point or an allow decision.

## Readiness boundary

This slice is `CONTRACT + EVALUATOR + SHADOW_ADAPTER`. Until a separate policy
adopts independently produced operational evidence, the fixed assurance
denominator remains `6/12`, and source-backed authority remains
`NOT_IMPLEMENTED / NONE`.
