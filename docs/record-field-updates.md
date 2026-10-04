# Build records one field at a time

Gooo bodies can update a declared Text field of a local record. The next
statement sees the current value. A saved record is a value copy; updating one
local leaves the other's fields intact. Saving a scalar field captures its
value at that statement.

```gooo
let copy = input0
let saved = copy
copy.title = input0.title
copy.state = "ready"
copy.reason = saved.reason + ":" + copy.state
return copy
```

Declare binary alternatives for the right-hand sides with `field_update`:

```gooo
choice "title" field_update at "0" alternative "input0.title" intent "Keep the original title."
choice "state" field_update at "1" alternative "\"ready\"" intent "상태를 ready로 만든다."
choice "reason" field_update at "2" alternative "copy.title + \":\" + copy.state" intent "Join the current title and updated state."
```

`at` counts field assignment right-hand sides in baseline source order, starting
at zero. Constructor choices use `field_value` with their own occurrence list.
The two kinds can share one assembly contract; adding an update does not change
an existing constructor ordinal. The nominal record and stable field ID come
from the typed receiver. A candidate replaces only the chosen expression.
Parameters stay read-only, and receivers must be existing local records.

Try the complete [sequential example](../examples/body-codegen/record-field-updates.gooo.fixture):

```sh
gooo body-codegen --json --activity Select examples/body-codegen/record-field-updates.gooo.fixture
gooo body-compose --source examples/body-codegen/record-field-updates.gooo.fixture \
  --cases examples/body-codegen/record-field-updates-cases.json --out /tmp/gooo-field-updates
```

With two deterministic attempts, the example matches9/15 selection fields.
With eight, it matches15/15. The separate native suite has14 named expectations
across two connected activities. Each run retains actual values and field IDs.
Saved source/selection replay reconstructs the same assignments with zero
predictions. An omitted model uses deterministic order.

An optional shared-field model ranks exactly three declared alternatives in one
prediction. Its current expression context contains complete field names,
ordered alternatives and intent. It observes field reads such as `copy.state`,
while reaching definitions, mutations and branch effects stay in the typed
source and evaluator. Independent field probabilities cannot measure the
completeness of a dependent program. Supplied cases measure that result after
each candidate executes. This representation opens the next small model
experiment: make current local definitions explicit in its input.

Records keep fixed sixteen-slot string-header storage in finite interpretation
and value structs in Go. No per-field mutable map is introduced. Existing limits
of six choices,64 attempts,128 cases and required Text fields still apply.
