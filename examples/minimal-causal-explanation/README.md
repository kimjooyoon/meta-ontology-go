# Minimal causal explanation

This is a read-only philosophy experiment for `meta-ontology-go`. Its claim is
narrow: a decision should expose the smallest sufficient causal path that would
change the decision when any member is removed. The path set is authoritative;
the explanation sentence is incidental prose and is never used by the judge.

## Formal choices

The experiment adopts two ideas from formal literature:

- [Halpern and Pearl, *Causes and Explanations: A Structural-Model Approach,
  Part II: Explanations*](https://arxiv.org/abs/cs/0208034) motivates grounding
  an explanation in actual cause and structural-model counterfactuals. Here
  that becomes a deterministic removal test: remove one evidence node and
  require the decision to change from `PASS` to `FAIL_CLOSED`.
- [Lynce and Marques-Silva, *On Computing Minimum Unsatisfiable
  Cores*](https://www.satisfiability.org/SAT04/programme/110.pdf) distinguishes
  a minimal core, where no member can be removed while retaining the property,
  from a minimum core, which globally optimizes cardinality. We adopt the
  deletion test and deliberately do not claim globally smallest paths.

For additional context, [Beckers, *Causal Explanations and
XAI*](https://proceedings.mlr.press/v177/beckers22a.html) separates sufficient
and counterfactual explanations. This experiment keeps that separation but
rejects natural-language generation, probabilistic ranking, and an
audience-dependent notion of usefulness as proof inputs.

## Fixed contract

The Go evaluator emits a sealed receipt and the `verify` package independently
judges it. The denominator is fixed at 12 indicators, 6 preserved claims, 12
append-only claim transitions, 3 cases, and 3 candidate paths:

| case | path | expected result |
| --- | --- | --- |
| `minimal` | request → policy → result | accepted: sufficient and minimal |
| `overlong` | request → policy → result + audit noise | rejected: sufficient but not minimal |
| `insufficient` | request → result | rejected: insufficient |

The receipt records 7 removal counterfactuals. Six are decision-changing: all 3
members of the accepted path and the 3 causal members of the overlong path.
Removing its audit noise leaves `PASS`, which is the constructive witness for
non-minimality. The available-evidence total is 11 across the three cases, but
the receipt does not enumerate those logs; only the authoritative path sets are
returned.

Every indicator records its producer, consumer, meta-operation, and proof
choice. Every claim transition records `stage/step/reason`, and transitions
preserve the claim (`UNRECORDED → OPEN → DISCHARGED`) instead of deleting it.
No repository workspace write, promotion, or semantic mutation is authorized.

## CI-only evidence

The workflow compiles this real `.gooo` source, generates two byte-identical
receipts, runs the independent judge, and exercises the positive, overlong, and
insufficient cases from the sealed receipt. It is intended to be verified by
GitHub Actions; no local test command is part of this experiment.
