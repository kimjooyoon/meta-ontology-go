# Caller-guided source IR search — 2026-10-08

Starting compiler source: 4a0cfdd8. Whole-program construction currently requires
record-choice bodies and rejects source-declared integer IR search. Connect the
existing finite expression grammar to caller-driven joint construction.

Freeze this example before implementation/model observations. The local zero
case accepts both `input` and `0`; Main requires the zero expression. Preserve
that local expectation and independently observe three caller inputs, including
an exact integer above 2^53.

- Source-owned grammar, retained candidates and per-activity attempt budget own
  admissible expressions. Caller feedback chooses within that bound.
- Keep typed search candidate evidence separate from record-choice masks, with
  explicit kinds and a versioned mixed-construction receipt.
- Preserve local counts/actuals, grammar truncation, original source identity and
  original model order. Replay re-derives every expression and local observation,
  then repeats native caller execution without loading a model.
- Exercise scalar-only, mixed record/search, explicit binding, conflicting local
  expectations, budget exhaustion and changed records in tests.
- Dogfood the unchanged graph model on record choices in a mixed program. Search
  expression ordering is deterministic. Do not attribute those choices to the
  graph model, retrain weights or change expected values to obtain a pass.
- Existing record-only receipts remain replayable. Source fill with several holes
  and scalar structural choices are further integration work.
