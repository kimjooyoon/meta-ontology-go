# Language value witness

This experiment adds one deliberately small value-level program to Gooo:
`Increment` selects the registered pure operation `int.add` with operand `1`.

The CI receipt records five exact input/output cases, eight fail-closed
counterexamples, three reader resolutions, and the fixed scoped coordinate
`0/1 -> 1/1`. The program participates in the bidirectional semantic
fingerprint. The lower core IR still rejects this program rather than silently
dropping it, so core IR preservation remains an explicit `0/1` non-claim.

This does not claim a general expression language, arbitrary value types,
runtime memory or performance bounds, or authority to mutate the repository.
