package main

const bodyConstructHelp = `Construct a program using caller feedback

` + bodyConstructUsage + `

Record choices, integer search and source_fill helpers keep their own Gooo cases.
Each completed combination is
compiled and run on --construction-cases; --cases runs after selection. The
optional --model ranks record choices; --fill-model uses a compatible operation
classifier for initial fill selection. Local fill holdouts stay separate from
training/caller scores and never decide completion. Without models,
the order is deterministic. Saved construction rechecks every attempted program
without inference. Consumed caller inputs are reported separately from new ones.
Use --format text for a short result or --format markdown for a shareable table.
The default --format json and saved --out files retain every original record.
Reports separate local checks, caller construction expectations and later evaluation;
missing observations stay unmeasured and failures keep the original exit status.
To read an existing full JSON result, use --report saved-result.json. This mode
defaults to text and accepts --format only; it needs no source, model or Go toolchain.
It reads the saved invocation without executing or reverifying it. Exit 0 means
the report was read, even when that invocation recorded a failure. --format json
returns the original file bytes. Save the full normal JSON stdout for this mode;
the separate --out construction.json or evaluation.json lacks the full envelope.
See examples/caller-guided-construction/README.md and examples/caller-source-fill/README.md.
`
