# Select an imported record type in an activity signature

Both left and right export an `Amount` record with a different stable ID. The
app's two source files bind the same Korean import alias to different packages.
Their input/result types resolve in each declaring file's import scope.

```sh
gooo package inspect --json workspace.json
```

The command requires development source containing the qualified-signature and
source-inspection changes. [The saved inspection](inspection.json) contains three
packages and records Main's input/result as `urn:qualified:left:amount`, while
Alternate's input/result use `urn:qualified:right:amount`. Four qualified signature
references match those IDs. One prepared call connects Main to left's `요금`.
The two same-named records have different lowered names and retain their IDs.

[The observation](observation.json) binds the actual source inspection to its
development binary, source overlays, module and output digests. The source fixture
was prepared once, with an empty PATH and unchanged inputs. Program executions,
model predictions and training runs are zero. Native values, runtime replay and
editor interactions were not observed.

The saved producer predates a later `maps.Copy` modernization and final package
compilation. Its exact source snapshot remains retained; the source inspection
was not replayed or rebound to a future accepted commit.

See [the language guide](../../docs/qualified-activity-types.ko.md).
