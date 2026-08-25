# Repository projection diagnostics

This diagnostic raises the resolution of the existing physical-storage guard without weakening it.

The fixed transition contract is:

- projection roundtrip loss: 0
- unbound manifest entries: 0
- observed physical directories above 10 direct children: 1
- classified GitHub workflow discovery roots: 1
- unclassified direct-entry subjects: 0
- blocking direct-entry subjects: 1 before classification, 0 after classification
- physical directories mixing branch and leaf entries: 0

Each topology violation must name its repository-relative physical path, observed value, limit, consumer, and meta-operation. The project root remains outside the nested storage topology rule.

Only `.github/workflows` containing regular `.yml` or `.yaml` files is `NOT_APPLICABLE` to generic radix sharding, with reason `GITHUB_WORKFLOW_DISCOVERY_ROOT`. Other directories above the limit remain blocking. The diagnostic Action validates the structured evidence and uploads the raw receipt without authorizing repository writes.
