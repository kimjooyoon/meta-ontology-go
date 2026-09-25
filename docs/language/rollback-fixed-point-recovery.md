# Rollback fixed-point recovery

## Problem

A guarded promotion can fail closed because its predecessor evidence is missing or
has a newer schema. If that failure also marks the producer workflow as failed,
the next run cannot select the producer and the cycle becomes self-referential.

## Meta operations

`project-promotion-contract` validates the current v2 proposal receipt and emits
one canonical v1 envelope for the unchanged guard. The projection is read-only,
has eight projected fields, zero field losses, zero repository writes, and no
mutation authority.

`recover-guarded-fixed-point` joins two exact receipts for one commit:

- guarded promotion is `FAIL_CLOSED / LOWER_RESOLUTION` with unresolved evidence;
- transformation is `FIXED_POINT / EXACT_FIXED_POINT`;
- effects, source-workspace mutations, repository writes, and mutation authority
  are all zero.

The result is `RECOVERED_FIXED_POINT`. It does not rewrite the guarded event to
`AUTHORIZED` and it does not authorize repository mutation.

## Quantified contract

The language denominator remains `24`. Registering the executable recovery
concept changes readiness from `11/24` to `12/24`, from `4583` to `5000` basis
points, with delta `+1` obligation and `+417` basis points. CI must report zero
regressions, zero unresolved readiness obligations, and zero repository writes.

## Use case

The first merged run after a bootstrap gap may preserve the guard failure while
closing at a fixed point. Its successful producer workflow then exposes the v1
compatibility envelope. The following merged run can be authorized by the
unchanged guard without circularly depending on its own workflow conclusion.
