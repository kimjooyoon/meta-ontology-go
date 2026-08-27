# Negative fixture contract

These target files are deliberately not observation evidence:

* `negative-duplicate-target.gooo` has two semantic `Root` occurrences and must
  fail with `TARGET_ACTIVITY_OCCURRENCE_CARDINALITY`/canonical reconstruction,
  never become a claim receipt.
* `negative-comment-only-target.gooo` contains only a comment-like Root and must
  fail closed because comments are not AST activities.
* `negative-invalid-target.gooo` has a syntax error and must fail with
  `TARGET_SYNTAX_OR_LOWER_INVALID`.

Structural inventory mutations are checked after resealing the bundle digest;
missing, duplicate, additional, and replacement rows must report their typed
`STRUCTURAL_INVENTORY_*` reason rather than passing on a top-level digest alone.
Procedure relabelling is similarly rejected by the fixed activity-to-procedure
contract even when a fixture is resealed.
