# Self-improvement observation

This example is a bounded, read-only compiler observation loop. The
`observation.gooo` authority declares the observed semantic activity, its
canonical normalized-IR identity, its allowed read effects, and the exact
duplicate-input candidate rule.

`gooo observe` runs the real `generate` lowering path twice on `input.gooo`.
It records the phase, operation ID, canonical input digest, and source span
for each evaluation. A candidate is emitted only when the same pure operation
is evaluated more than once for the same digest. Safety and benefit are left
unknown; no repository change is adopted.

The CI loop binds that report into the existing six-artifact semantic
operation envelope for NORMAL, UNKNOWN, REFUTED, and REPLAY cases, then
independently re-verifies the envelope and the observation decision.
