# Readiness push continuity

## Observed gap

Merge commit `37179a5d58bd00bd169cf8e997bf444dc755c6c2` completed its
push CI successfully, but readiness-dependent workflows were not selected:

| exact observation | count |
| --- | ---: |
| metric strategy push runs | 0 |
| transformation workflow runs | 0 |
| readiness artifacts for the merge SHA | 0 |

The metric strategy push filter did not include any language-readiness path.
This is an execution-topology gap, not a readiness regression.

## Fixed contract

Five dependency path classes must select metric strategy generation:

1. Transformation workflow definition.
2. Guarded-promotion witness.
3. Language-readiness witness.
4. Language-concept meta model.
5. Language-readiness meta model.

Each class must occur exactly once in both `pull_request` and `push` filters.
The fixed conformance denominator is therefore `5 * 2 = 10` coordinates.
The Go verifier fails unless the result is exactly `10/10`.

## Acceptance receipt

This foundation change claims `10/24 -> 10/24`, completed delta `0`, and basis
point delta `0`. Its post-merge acceptance values are:

| outcome | target |
| --- | ---: |
| metric strategy push run for the merge SHA | 1 |
| successful strategy job | 1 |
| transformation workflow run for the merge SHA | 1 |
| readiness artifact for the merge SHA | 1 |
| readiness fixed point | `10/24` |

No manually dispatched run counts because its event identity differs from the
required merged `push` event.

## Use cases

- Regenerate readiness after changing its evaluator, concept model, or witness.
- Regenerate downstream readiness when promotion or transformation policy changes.
- Prevent a PR-only artifact from being mistaken for a merged-commit baseline.
