# Toolchain CLI readiness

## Fixed denominator

The language readiness registry remains `24` obligations. This change can
satisfy only `TOOLCHAIN-CLI`; it cannot satisfy format/fix, LSP, conformance, or
cross-platform release.

The only accepted transition is:

| Measure | Before | After | Delta |
| --- | ---: | ---: | ---: |
| completed obligations | 19 | 20 | +1 |
| readiness basis points | 7916 | 8333 | +417 |
| regressions | 0 | 0 | 0 |
| unresolved evidence | 0 | 0 | 0 |

## Indicators

The fixed report has `18` indicators: `3` OUTCOME, `7` DRIVER, and `8`
GUARDRAIL. Targets are twelve satisfied cases, twenty-four invocations, thirteen
declared commands, three structured outputs, four language operations, twelve
identical replays, one binary digest binding, and zero mismatches, writes,
mutation authority, registry drift, or unresolved observations.

The twelve-case registry denominator is unchanged. The declared-command driver
changes from eleven to thirteen because `format` and `fix` are now visible;
that driver change is not counted as another readiness obligation.

## Munchhausen choices

`FOUNDATION` binds the exact head, concept artifact, registry, Go 1.27 runtime,
and executable digest. `COHERENCE` checks the command and structured-output
contracts twice. `REGRESSION` proves all six rejection paths and all zero-value
effect guardrails.
