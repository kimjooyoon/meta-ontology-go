# Original CI failure and source correction

The original pull-request CI run [37977164250](https://github.com/kimjooyoon/meta-ontology-go/actions/runs/37977164250)
completed with FAILURE at source `3544519e79eda26ba6d3a167427bbd22ace7e040`, attempt 1.
Format, vet, unit tests, race tests and policy passed. Semantic conformance failed;
the dependent evidence and proof jobs therefore failed. The original readiness
run `37977164217` passed all four native platforms. Neither run was restarted.

The extractor reported `FREE_BINDINGS_UNPROVEN` for
`RealizeSourcePathCandidate`: its generated helper rendered to 78 lines,
three over the existing 75-line capacity. The source function itself had 67 lines.
The raw semantic job log preserves that failure and its original source digest.

The correction moves the local-case scoring into `scoreSourcePathCandidate`.
It keeps candidate ordering, binding, evaluation and rejection behavior intact.
The corrected source file has SHA256
`e48dc21c81f9c2653bbe8b224de5fa04e6a4a991670a2d214cb83490194b871a`.
Focused source-path and joint race checks passed (1.670 and 17.432 seconds),
and vet passed. The same extractor produced five files with PASS; their
generated Go passed race tests through a read-only overlay (1.620 seconds).
These are local correction observations, pending a new exact-source original CI.

Each compressed file was byte-compared with its uncompressed original.
The earlier clean-producer observations and their manifest remain unchanged.
