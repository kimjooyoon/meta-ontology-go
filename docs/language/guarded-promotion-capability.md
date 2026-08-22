# Guarded promotion capability

## Semantic boundary

A promotion event and a promotion capability are different typed claims.
The event receipt keeps its observed decision. A failed event is never skipped,
rewritten, or counted as authorization. The capability receipt asks whether the
exact implementation that produced one real authorized event remains present.

## Versioned foundation

The FOUNDATION is pull request 399, merge subject
`027785b635246537354ac7517054b755235ae765`. GitHub Actions run `32599122103`
produced artifact `9482381929` with archive digest
`sha256:6fe9223adb6260e644bba5cfb28247291623ba139c4773bb15da139ae3da63f0`.
Its report is `AUTHORIZED`, `12/12`, `10000 BPS`, zero unresolved evidence,
zero repository writes, and no mutation authority.

CI verifies the external archive when this foundation is introduced. Later
replays use the byte-identical versioned report and require the foundation to
be an ancestor of the current subject. The guarded-promotion package tree and
its witness tree must both equal their foundation Git trees.

## Indicators

The receipt has exactly eight coordinates, eight indicators, and three proofs.
Indicators are split into one outcome, three drivers, and four guardrails.
Unknown ancestry or tree evidence lowers resolution. Known tree drift rejects
the capability. Either result grants zero readiness obligations.

## Exact readiness use case

The fixed registry remains 24 obligations. Given a comparable predecessor at
`10/24` and `4166 BPS`, an exact capability receipt may satisfy only
`AUTONOMY-GUARDED-PROMOTION`. The only accepted gain is `11/24`, `4583 BPS`,
`+1` obligation, and `+417 BPS`, with zero regressions and zero unresolved
evidence. These are acceptance values, not a claim made before CI observes them.

## Preserved failure use case

The merge of pull request 402 observed strategy evidence, a transformation
trigger, and direct CI, but no readiness artifact: exactly `3/4`. Its current
promotion event remained `FAIL_CLOSED / GUARDED_PROMOTION_EVIDENCE_UNKNOWN`
with `10/12` coordinates. The capability receipt must not alter that event.
