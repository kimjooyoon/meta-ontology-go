# Operation interface experiment

`gooo emit --kind operation-interface --entry PayOrder examples/billing-package`
projects the same Gooo package receipt as the operation manifest, but removes
source definition receipts from the public artifact.

The interface is a deterministic consumer-facing JSON projection. Its
`INTERFACE_ONLY` resolution is intentional abstraction, not a claim that the
compiler knows less about the package.

The experiment does not claim interface completeness, compatibility stability,
production readiness, or general-purpose generation.
