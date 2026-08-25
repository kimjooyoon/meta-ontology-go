# Repository projection diagnostics

This diagnostic raises the resolution of the existing physical-storage guard without weakening it.

The fixed current-state contract is:

- projection roundtrip loss: 0
- unbound manifest entries: 0
- physical directories above 10 direct children: 1
- physical directories mixing branch and leaf entries: 0

Each topology violation must name its repository-relative physical path, observed value, limit, consumer, and meta-operation. The project root remains outside the nested storage topology rule.

The diagnostic Action treats the projector's expected fail-closed exit as an observation, validates the structured evidence, and uploads the raw receipt. It does not add an exception, move a path, authorize a write, or claim improvement.
