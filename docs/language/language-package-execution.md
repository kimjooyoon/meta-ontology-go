# Language package execution

`gooo run --entry PayOrder examples/billing-package` treats the immediate `.gooo` files in the directory as one semantic package. Files are sorted by canonical relative filename, parsed independently, checked for one package and namespace, combined as syntax declarations, and then lowered through the existing single-source execution boundary.

## Fixed denominator

The CI contract contains exactly 5 cases: positive execution, deterministic replay, package-header rejection, duplicate-declaration rejection, and source-count rejection. The denominator changes only by changing the versioned contract and its meta reducer together.

The human-visible indicators are exact counters:

| Indicator | Target | Meaning |
| --- | ---: | --- |
| `PACKAGE_FIXED_CASES` | 5/5 | Every fixed case has its declared decision and reason. |
| `PACKAGE_SOURCE_FILES` | 2/2 | The positive package binds two source files. |
| `PACKAGE_EXECUTIONS` | 1/1 | One distinct positive package execution succeeds. |
| `PACKAGE_DETERMINISTIC_REPLAYS` | 1/1 | Replaying the same package produces the same receipt digest. |
| `PACKAGE_DIAGNOSTIC_REJECTIONS` | 3/3 | All fixed invalid inputs fail closed. |
| `PACKAGE_EVENTS` | 7/7 | Two parse, one package-bind, and four execution events are visible. |
| `PACKAGE_UNKNOWN_DECISIONS` | 0/0 | Unknown top decisions lower resolution and fail closed. |
| `PACKAGE_REPOSITORY_WRITES` | 0/0 | Execution writes no repository file. |
| `PACKAGE_MUTATION_AUTHORITIES` | 0/0 | Execution acquires no mutation authority. |

## Reader-dependent resolution

`USER` sees outcome facts, `TOOL_AUTHOR` also sees operational binding and replay facts, and `GOVERNOR` sees every proof and effect fact. The visible set changes, but all views carry the same `facts_digest`; a lower-resolution view cannot invent a different result.

## Proof choices

The Munchausen choice is explicit per claim: `FOUNDATION` fixes the denominator and rejects unknown top decisions, `COHERENCE` binds parser, package, semantic, and effect receipts, and `REGRESSION` compares deterministic replay digests. No proof choice authorizes repository mutation.

If the new `run-package` journey passes in CI, the fixed delivery coordinates move from `31/36` to `32/36`, `USER 9/12` to `10/12`, cumulative `TOOL_AUTHOR 19/24` to `20/24`, and cumulative `GOVERNOR 31/36` to `32/36`. These are contract transitions, not performance-improvement claims.
