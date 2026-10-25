# Source subject witness

This metaprogram binds every emitted source observation to an explicit
canonical witness. It consumes the CI-owned source metrics report rather than
walking the checkout again.

The ledger distinguishes three spaces:

- `LOGICAL_FILE` records every file and its Go, Gooo, or other line count.
- `LOGICAL_DIRECTORY` records projected folder and file observations.
- `STORAGE_DIRECTORY` records policy-bearing physical topology observations.

Go and Gooo files must bind one-to-one to their existing source indicators.
Storage directories must bind all six topology indicators. Logical directory
and other-file observations are marked as derived instead of pretending an
upstream indicator exists.

The project root is measured but remains exempt from topology enforcement.
No root `README.md` is required. The exception, all subject witnesses, and the
complete upstream meta-indicator ledger are included in the semantic digest.

The companion Actions workflow builds the ledger twice, verifies each output,
and requires byte-identical canonical replay before publishing the artifact.

Nested Go module ownership is retained from the source report in file and function
witnesses, source observations and their digests. File/function indicator owners
must match the owning file observation; a disagreement returns no ledger. This
consumer reads the recorded report and does not rediscover module ownership from
the current checkout. Unknown fields still refuse decoding with
`SOURCE_SCHEMA_UNSUPPORTED` and `update-source-metrics-consumer`; invalid JSON
remains `SOURCE_JSON_MALFORMED`. Existing rewrite applicability rules are unchanged.

The regression uses the actual line-metrics producer and meta-binding constructor
on a small nested-module repository, followed by this consumer. Independent
expected line counts, source paths, unresolved observations and owner-tampering
cases check the connection across those components.

The original [workflow failure](https://github.com/kimjooyoon/meta-ontology-go/actions/runs/40164129378)
consumed source commit `68867ca2e82dd6d08f3237dc58ec99852f7c8c88` from CI run
`40162319878`. Its unchanged 38,201,544-byte report now produces and verifies
23,515 witnesses for 41,060 indicators, with byte-identical replay. The 26
module-owned witnesses and three unsatisfied module observations remain present;
the claim stays `OPEN`. This is a repaired consumption of that historical input,
not a rerun or replacement of the original CI result.
