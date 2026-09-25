# Guarded readiness promotion

## Boundary

`AUTONOMY-GUARDED-PROMOTION` does not grant repository mutation authority.
It grants readiness credit only after GitHub reports a successful `CI` run for
the default branch and one exact merged predecessor promotion receipt exists.

## Fixed contract

The receipt has 12 coordinates, 8 indicators, and 3 Munchhausen proofs.
The denominator is versioned by `gooo/autonomy-guarded-promotion/v1`.

| Source state | Satisfied | Total | Basis points | Decision |
| --- | ---: | ---: | ---: | --- |
| Pull request CI | 10 | 12 | 8333 | `DENIED` |
| Merged default-branch CI | 12 | 12 | 10000 | `AUTHORIZED` |
| Unknown or ambiguous evidence | exact count | 12 | integer result | `FAIL_CLOSED` |

The pull request state fails only the merged-push event and default-branch
coordinates. Unknown events lower semantic resolution instead of becoming a
fixed point or an authorization.

## Indicator split

The outcome is promotion readiness basis points. Drivers are the unique
predecessor count and source-integrity basis points. Guardrails count unmerged
boundary debt, ambiguous predecessors, unresolved evidence, observer writes,
and repository mutation authority.

## Trilemma choices

`FOUNDATION` binds the GitHub artifact identity, artifact digest, file digest,
report digest, and predecessor SHA. `COHERENCE` cross-links the CI workflow,
subject SHA, and successful conclusion. `REGRESSION` rejects pull-request
promotion, non-default branches, observer writes, and mutation authority.

## Executable use case

After a pull request is merged, `Guarded promotion conformance` observes the
completed default-branch `CI` run, selects the exact predecessor Transformation
artifact, downloads its promotion report, and emits the receipt twice. Byte
inequality, a missing candidate, multiple candidates, or a non-successful
source prevents authorization.

This foundation receipt does not change the 24-obligation readiness numerator.
A later consumer may move `10/24` to `11/24` only from an exact merged
`12/12 AUTHORIZED` receipt with zero unresolved evidence and zero writes.
