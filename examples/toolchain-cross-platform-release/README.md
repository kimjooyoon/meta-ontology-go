# Cross-platform release corpus

This v1 corpus fixes a denominator of three native x64 targets and twenty cases.

- Linux uses `ubuntu-24.04` and emits `tar.gz`.
- macOS uses `macos-15-intel` and emits `tar.gz`.
- Windows uses `windows-2025` and emits `zip`.

Each runner builds twice with Go 1.27.0, runs `gooo version --json` natively,
packages twice with fixed metadata, and emits one source-bound receipt.

The aggregate witness accepts only three unique `PASS / EXACT` receipts. Missing,
duplicate, stale, dirty, unresolved, or unknown evidence fails closed.
