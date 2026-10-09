# Cross-platform release corpus

This v1 corpus fixes a denominator of four native targets and twenty-six cases.

- Linux uses `ubuntu-24.04` and emits `tar.gz`.
- macOS uses `macos-15-intel` and emits `tar.gz`.
- macOS arm64 uses `macos-15` and emits `tar.gz`.
- Windows uses `windows-2025` and emits `zip`.

Each runner builds twice with Go 1.27.2, runs `gooo version --json` natively,
packages twice with fixed metadata, and emits one source-bound receipt.

The aggregate witness accepts only four unique `PASS / EXACT` receipts. Missing,
duplicate, stale, dirty, unresolved, or unknown evidence fails closed.

The 0.6.11 platform witness additionally constructs and replays division (eight
cases), retry (twelve) and filename classification (twelve) with the candidate
binary. It consumes the filename source-only v3 graph export through the SDK.
It also executes three input-only filename workspace requests and saved replay,
checking their actual fields and zero new predictions without assigning the
requests a supplied-expectation score.
It also runs the three-package diagnostic with both fully specified and
source-derived workspace metadata. Each form constructs and replays the same
four input rows and eight named expectations. All four runs must retain the
same generated program, match every independent expected output and make zero
new runtime/replay predictions. Four raw package receipts are retained per
platform; repeated executions do not increase the distinct-input count.
These observations remain separate from the twenty-six structural corpus cases;
they do not change the corpus denominator or measure trained model quality.

The 0.6.21 native profile also constructs the scalar identity example with Korean
and English type names, fixed selection and the included unchanged own graph model.
It checks 12/12 authored fields over four selection cases and four different native
input tuples, with exact large integers, Boolean and text fields. All four modes
select identical native code; each saved replay performs zero new inference.
The profile retains eight raw construct/replay reports per platform. The own model
must actually infer once during construction and retain its metadata/weight hashes.
These supplied-program checks remain separate from the structural corpus denominator.
