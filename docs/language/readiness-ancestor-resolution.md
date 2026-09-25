# Readiness ancestor resolution

## Decision

Readiness comparison may cross a missing artifact only when the exact selector
returns `FAIL_CLOSED / READINESS_PREDECESSOR_NOT_FOUND`. Every other result stops
resolution. The resolver never changes an artifact and reports repository writes
as `0`.

The search limit is the schema constant `8`. Attempts form a contiguous,
single-parent chain. The first exact selected artifact ends the search, so a more
distant artifact cannot replace valid or malformed closer evidence.

## Evidence incident

The current ancestry has three inspected coordinates:

| depth | commit | canonical readiness pair | decision |
| ---: | --- | ---: | --- |
| 0 | `f72c362c489e5ace515cec6093adf47faf1276e4` | 0 | skip only as not found |
| 1 | `027785b635246537354ac7517054b755235ae765` | 0 | skip only as not found |
| 2 | `75b892f4bc661b4a58a5517532c44662dae6eedf` | 1 | select |

The selected readiness artifact is GitHub artifact `9481822718`, digest
`sha256:bc9b6f46097a5876d87a185f73711415e410f42cad755d0bcf783970a106c0b6`.
Its binding artifact is `9481822194`, digest
`sha256:22efdf9d809c1e033a54c6d2c6168a992d4dcbb191ae16b5678822d4df96050f`.
The selected denominator is fixed at `24` and the completed count is `10`.

## Quantified meaning

This foundation change claims no readiness gain: `10/24 -> 10/24`, delta `0`.
Its acceptance receipt has a fixed `10`-coordinate denominator:

| kind | coordinates |
| --- | ---: |
| outcome | 2 |
| driver | 4 |
| guardrail | 4 |
| total | 10 |

Success requires `10/10`, three passed Munchhausen proofs, ambiguity `0`, writes
`0`, and readiness-delta claims `0`. These numbers restore an exact comparison
boundary; they do not count as language capability completion.

## Use cases

- Recover comparison continuity after squash or administratively merged commits
  whose workflows did not publish a readiness pair.
- Stop at expired, invalid, ambiguous, failed, unbound, or writing evidence rather
  than hiding it behind an older successful artifact.
- Replay historical selection byte-for-byte while preserving the selected
  artifact and binding bytes unchanged.
