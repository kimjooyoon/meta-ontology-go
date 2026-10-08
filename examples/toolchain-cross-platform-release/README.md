# Cross-platform release corpus

This v1 corpus fixes a denominator of four native targets and twenty-six cases.

- Linux uses `ubuntu-24.04` and emits `tar.gz`.
- macOS uses `macos-15-intel` and emits `tar.gz`.
- macOS arm64 uses `macos-15` and emits `tar.gz`.
- Windows uses `windows-2025` and emits `zip`.

Each runner builds twice with Go 1.27.1, runs `gooo version --json` natively,
packages twice with fixed metadata, and emits one source-bound receipt.

The aggregate witness accepts only four unique `PASS / EXACT` receipts. Missing,
duplicate, stale, dirty, unresolved, or unknown evidence fails closed.

The 0.6.9 platform witness additionally constructs and replays division (eight
cases), retry (twelve) and filename classification (twelve) with the candidate
binary. It consumes the filename source-only v3 graph export through the SDK.
These observations remain separate from the twenty-six structural corpus cases;
they do not change the corpus denominator or measure trained model quality.
