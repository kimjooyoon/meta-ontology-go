# Incremental typed paths

The optional local model judges bounded Korean or English intentions for Gooo
structural alternatives. A session ranks those alternatives once, then evaluates
new candidate bodies in batches without repeating earlier masks. It does not
train, rewrite source files, call an external provider, or rerank from test/CI
feedback. Changing the plan, cases, seed or model requires a new session.

```sh
gooo body-codegen --json --path-plan plan.json --path-model model.json \
  --path-step-attempts 8 --activity ConditionalAssign source.gooo
```

Omit `--path-model` for deterministic declared-fallback ordering. Omit
`--path-step-attempts` for the existing search. The document retains its total
1..64 candidate budget and 1..128 finite cases. Each step evaluates 1..64 new
candidates. Model ranking happens after binding the fallback to authoritative
Gooo source and before candidate tests. The compiler emits and verifies the
final selected body; the JSON receipt includes initialization and batch progress.
This CLI returns one final JSON result, not streaming native Go emissions.

The receipt separates:

- Typed lowering completeness from finite functional completeness. A body can
  lower completely while satisfying only six of seven declared cases.
- Cumulative evaluated and type-rejected candidates from unattempted masks.
- Initial model calls from zero new model predictions during each advance.
- The eight-byte scheduled-mask bitset for 64 alternatives from frontier,
  prepared body, model, receipt and whole-process memory.

Progress hashes link observations; they do not authorize edits or establish
ontology facts. Source binding, stable activity identity, deterministic replay,
native type checking and parity with the typed interpreter remain required.
Partial outcomes are retained with their explicit finite denominator.

The Go SDK `v0.2.2-experimental` exposes `NewSession`, `Observe`, `Advance` and
`SearchBatches`. Its core session keeps one best body, a frontier and a bitset,
without previous attempt logs. The native compatibility helper retains up to
64 attempts and their batch observations for its final receipt. The standalone
research CLI can stream these observations across the larger declared finite
space; native documents continue to cap total attempts at 64.

Concurrent calls on the same SDK session fail immediately with `ErrSessionBusy`.
Candidate evaluation checks context cancellation and requeues an unfinished
candidate without recording a partial case prefix. Generic output writers can
still block outside the evaluator: process harnesses must bound and drain
output and enforce a process deadline. This is not a guarantee of all-input
correctness, general language accuracy or arbitrary I/O deadlock freedom.
