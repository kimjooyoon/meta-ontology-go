# Track inputs across body construction and execution

Gooo first fills the two holes in `Lift`, then executes the `Normalize → Lift`
graph. The input-separation observation includes both that earlier construction
and the values actually delivered during native execution.

```sh
gooo package execute --json \
  --cases examples/workspace-input-observations/cases.json \
  examples/workspace-input-observations/gooo.workspace.json
```

Pass `--go /path/to/go1.27.1` when the Go executable on PATH differs from the
required toolchain. This example runs locally with deterministic construction.

| Root input | Input delivered to Lift | Earlier observation | Classification |
| ---: | ---: | --- | --- |
| -9 | 0 | Construction case | Overlapping |
| 11 | 11 | None in the recorded construction inputs | Disjoint |
| 100 | 100 | Holdout case | Overlapping |
| 11 | 11 | Duplicate root input | Counted with the previous 11 |

Four rows specify eight activity outputs. The whole-case measurement has three
unique roots, one duplicate, two overlapping cases and one disjoint case. Its
finite fraction is 1/1 when all expectations for input 11 match. A conflicting
expectation on its duplicate changes that fraction to 0/1 and `PROGRESS`.

`result.runtime.input_separation.earlier_stages` binds each earlier body fill to
its activity, original source, plan and canonical input-set digest. Its inputs
include construction cases, holdout cases and generated behavioral probes.
The actual tuples remain in `result.body_fills`. Integer vectors and typed
record tuples retain their port order and exact integer values.

The same measurement supports source-declared fills and externally supplied
fill plans. Missing input evidence remains `UNKNOWN`. Input-only executions
keep their overlap counts but have no expected-output success count. Exposure
in a model's training data remains unknown.
