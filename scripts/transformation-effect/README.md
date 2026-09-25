# Transformation effect

`transformation-effect` is the non-authorizing execution layer between source
metrics and a proposed repository change.

It consumes the exact-head source metrics, generation plan, execution manifest,
receipt report, and provenance envelope emitted by CI. The tool replays their
existing contracts before doing any work.

For a fixed point it proves that a disposable worktree has identical before and
after trees and emits an empty content patch. For a plan it runs only registered
executors against exact indicator subjects, remeasures the sandbox, rejects any
residual actionable subject, seals evaluator receipts, and emits changed content
with before/after digests.

The source workspace is hashed before and after execution. All writes are
restricted to a disposable Git worktree. The resulting ledger, receipts,
provenance, and patch never authorize promotion.

Project-root topology remains not applicable. Root counts are evidence, not a
requirement for a root `README.md`.
