# Inspect source IDs through an imported call

The app imports `example/core` as `rules` and calls `rules.요금(input)`. Both the
entry and helper declare stable activity IDs. The source inspection shows those
IDs in the original and lowered call participants.

```sh
gooo package inspect --json workspace.json
```

The command reads and prepares source. It leaves the files unchanged and makes
zero model predictions and program executions. See
[the fields and coverage](../../docs/workspace-source-inspection.md).

## Recorded source inspection

[The saved inspection](inspection.json) follows one imported call between two
packages. The two participants keep `urn:workspace:entry` and
`urn:workspace:charge` in both identity fields and the prepared source.
[The observation](observation.json) retains the actual development binary and
source overlays. Files remained unchanged; no native body or model was run.

The [initial fixture](declined-qualified-types/observation.json) used the
unsupported `rules.Integer` spelling in a signature and was declined. Its source
and original response are retained. The valid example uses the visible imported
entity name while keeping the qualified Korean activity call. Each distinct
source fixture was inspected once. The binary was built before the later help
text addition; its recorded producer identity is retained.
