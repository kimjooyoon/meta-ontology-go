# Saved transport report origins

This example reads a saved `prooftransport.Report` and connects its content to
the report's exact bytes, JSON locations, and original Gooo definitions. The
optional view makes no registry request and does not change the execution path.

```go
view, err := transportorigins.FromSavedReport(raw)
```

The source in `origins.gooo` owns the consistency decisions: compare the reported plan,
initial selection, readback selection, and outcome with the existing Gooo
definitions; retain unresolved observations when they disagree. Its zero-input
`ReportByteLimit()` stage supplies the byte bound before JSON decoding and also
binds that value into the generated composition. No caller value stands in for
this source-owned constant.

The adapter reads the native report's existing Go JSON tags. Exact field names
own their locations. Unknown extensions remain in the original bytes used for
content identity; case-folded aliases cannot overwrite a canonical field.
Missing required fields produce an unresolved view. Invalid JSON, duplicate
keys, invalid UTF-8, and input above the source-defined 256 KiB bound are errors.

A report identifier hashes the exact raw bytes. Record identifiers combine that
content identity with a role, and `recordPath` identifies its canonical JSON
location. The RDF statements describe entities and proposed plans. The view
marks the execution identity as `UNOBSERVED`: identical report bytes can come
from different invocations, and a recorded readback can exist even when
cancellation prevented a network dispatch. A proposed next operation stays
`PROPOSED`.

Reported generation-context identifiers hash the decoded UTF-8 string values in
`generation_origin` and `outcome_generation_origin`. They identify reported
metadata. The view's own `GoooGenerationOrigin` separately records the compiler
and nine source files that produced this projection. A concrete projection
identifier binds that immutable context as well as the report bytes. It
specializes the generic report view, so changing the view generator cannot
silently merge concrete projections. This uses the fixed-aspect distinction in
[PROV-DM specialization](https://www.w3.org/TR/prov-dm/#term-specialization);
it does not assert a unique physical execution.

To regenerate from the repository root using a compiler that supports zero-input
composition stages:

```sh
gooo package generate --go-package transportorigins --embed-origin \
  --export-stage ReportByteLimit=gooo://transport-origins/report-byte-limit \
  --out examples/ci-strategy/transport-origins \
  examples/transport-origins.workspace.json
```

The checked-in projection was produced by clean development compiler
`ba353f5af2e94389cae3af05470ae09e83109141`. Its recorded producer remains that
revision when this example is delivered in a later commit.

Functional tests run the production transport runner with controlled HTTP
responses, then read its serialized report. They cover read-only observation,
successful storage, a failed response with an observed effect, an unobserved
effect, and cancelled readback. The view must leave bytes and request counters
unchanged. Contradictory records yield no statements; a retained older report
with missing later fields stays unresolved. These finite cases do not establish
remote registry availability, invocation identity, or an RDF reasoner's results.

The source package separates three responsibilities: `origins.gooo` admits a
consistent report, `render.gooo` describes its RDF view, and `terms.gooo` owns
bounded term formatting. The Go adapter stays unchanged. A rendering definition
can be read separately from the report-admission logic; the package loader keeps
their exact source origins and stable activity IDs together.
