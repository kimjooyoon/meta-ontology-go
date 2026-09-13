# Language value witness

This experiment adds one deliberately small value-level program to Gooo:
`Increment` selects the registered pure operation `int.add` with operand `1`.

The CI receipt records five exact input/output cases, eight fail-closed
counterexamples, three reader resolutions, and the fixed scoped coordinate
`0/1 -> 1/1`. The program participates in both the bidirectional and core IR
semantic fingerprints. Core IR preservation and fingerprint sensitivity are
each `1/1`; an unknown declaration attribute remains fail-closed at `1/1`.

Its receipt scope is `REGISTERED_VALUE_OPERATION`, which names the declared
value-evaluation boundary. The scope field alone does not prove that an
`Apply` call completed: the passing report's invoked-operation count, outputs,
and result evidence establish that narrower fact. This does not claim
handwritten Go-body execution or external effects. Historical v2 reports with
the same schema name but no scope are rejected rather than defaulted.

This does not claim a general expression language, arbitrary value types,
core IR execution or code generation, runtime memory or performance bounds, or
authority to mutate the repository.
