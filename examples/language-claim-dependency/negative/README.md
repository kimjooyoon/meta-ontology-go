# Negative fixture contract

These target files are deliberately not observation evidence:

* `negative-duplicate-target.gooo` has two semantic `Root` occurrences and must
  fail with `TARGET_ACTIVITY_OCCURRENCE_CARDINALITY`/canonical reconstruction,
  never become a claim receipt.
* `commented-out-activity.gooo` comments out the Root activity, leaving a
  non-complete semantic inventory; it must fail closed because comments are not
  AST activities.
* `negative-invalid-target.gooo` has a syntax error and must fail with
  `TARGET_SYNTAX_OR_LOWER_INVALID`.

Structural inventory mutations are checked after resealing the bundle digest;
missing, duplicate, additional, and replacement rows must report their typed
`STRUCTURAL_INVENTORY_*` reason rather than passing on a top-level digest alone.
Procedure relabelling is similarly rejected by the fixed activity-to-procedure
contract even when a fixture is resealed.

`../comment-whitespace-target.gooo` is the separate valid comment/whitespace-only
fixture: raw bytes and span provenance change, while canonical semantic addresses,
per-occurrence semantic digests, transitions, and decision remain unchanged.
