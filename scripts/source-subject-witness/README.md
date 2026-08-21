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
