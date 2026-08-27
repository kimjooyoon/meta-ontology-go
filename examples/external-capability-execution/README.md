# External capability execution

This use case separates two statements that must never be merged:

- the pinned gomacro repository is not fully compatible with Go 1.27: `6/8`, `FAIL_CLOSED`
- selected metaprogramming capabilities can still execute exactly without changing that parent result

The witness builds the pinned commit with Go 1.27.0 in CI. It executes embedded
evaluation and AST macro generation twice, compares normalized receipts, and
requires both the Gooo checkout and the gomacro checkout to remain unchanged.
Temporary build and generated files live outside both repositories.

## Fixed metric denominator

| # | Class | Munchausen choice | Meta-operation | Exact target |
| --- | --- | --- | --- | --- |
| 1 | driver | FOUNDATION | bind pinned repository | `1/1` |
| 2 | driver | FOUNDATION | verify pinned commit | `1/1` |
| 3 | driver | FOUNDATION | verify pinned tree | `1/1` |
| 4 | driver | FOUNDATION | select Go toolchain | `go1.27.0` |
| 5 | outcome | COHERENCE | execute embedded evaluation | `42` in `2/2` runs |
| 6 | outcome | COHERENCE | execute interpreted function | `55` in `2/2` runs |
| 7 | outcome | COHERENCE | execute AST macro generation | pinned output in `2/2` runs |
| 8 | guardrail | COHERENCE | compare normalized replay | `1/1` equal pair |
| 9 | guardrail | REGRESSION | preserve repository boundary | `0` changed paths |
| 10 | guardrail | REGRESSION | preserve parent failure | `FAIL_CLOSED / EXACT / 6/8` |

Only `10/10` produces `CAPABILITY_EXECUTABLE / EXACT`. This decision has
`NO_EFFECT`, `official_mutation_count=0`, and `promotion_count=0`; it cannot
raise language assurance above `11/12` or rewrite whole-project compatibility.

## Resolution behavior

The regression suite has a versioned denominator of `15`: one exact case,
three unknown-evidence cases, and eleven known invariant violations. Unknown
run states and unknown parent decisions become `FAIL_CLOSED / UNKNOWN` rather
than a fixed point. Known mismatches become `FAIL_CLOSED / INVARIANT_ONLY`.

The positive concept is capability-scoped external evidence: a language can
retain useful, executable metaprogramming evidence without claiming that an
entire upstream project works. This is an uncommon combination claim, not a
novelty claim. CI is the execution authority; local test executions remain zero.

The project root remains exceptional: root README and root topology are
`NOT_APPLICABLE` and are not included in this capability denominator.

The `authorization` subdirectory adds a second boundary. Executability remains
evidence, while a Gooo policy and a default-deny checker decide whether the
requested operation is authorized. The bootstrap receipt intentionally remains
`9/10 / UNKNOWN` until a merged CI artifact supplies an external policy
foundation; a green workflow does not rewrite that semantic result.
