# Toolchain executable use cases

This directory is a versioned input to `toolchain-usecase-witness`, not a list
of narrative examples. The evaluator binds the registry to the exact checked
out language-concept artifact and executes all three cases.

| Case | Input mutation | Required decision |
| --- | --- | --- |
| canonical concept artifact | none | `PASS` |
| replay digest tamper | replace replay digest | `FAIL_CLOSED` |
| repository write tamper | set writes to one | `FAIL_CLOSED` |

The fixed denominator is three. Missing, reordered, renamed, or unknown cases
produce `FAIL_CLOSED / LOWER_RESOLUTION`; a known decision mismatch produces
`FAIL_CLOSED / EXACT`. The witness writes only its receipt outside the source
tree and grants no mutation authority.
