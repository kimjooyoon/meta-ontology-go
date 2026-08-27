# Phase separation witness

This experiment uses the repository Gooo grammar: every fixture is a real
`package`/`namespace` file with `entity` declarations and `activity ...
computes` programs. The computes payload is structured data, not a second DSL:
it names phase-local value IDs and literal classes, the source/target phases,
and the transfer payload class.

`main.gooo` contains one clean claim with two explicit adjacent transfers.
`leaks.gooo` contains five source-derived counterexamples: value, authority,
evidence, a source-to-execution phase skip, and an execution-to-expansion phase
reverse. `unknown.gooo` is syntactically valid Gooo but has an unsupported
case coordinate, so the claim remains `OPEN` at `SOURCE/PARSE/UNKNOWN_SOURCE_CONTRACT`.

The producer and consumer each call `syntax.ParseFile` followed by
`bidir.Lower`. The consumer has its own wire model and evaluator and imports no
producer package. It reconstructs outcomes from endpoint adjacency, phase-local
IDs, payload class, claim digest, target declaration digest, and evidence
provenance. Receipt expected/actual labels and counts are not adjudication
authority.

The receipt carries append-only previous evidence digests, claim lifecycle
states, three audience views, twelve derived indicators, semantic and comment-only
interventions, and an independently observed CI read-only snapshot.
