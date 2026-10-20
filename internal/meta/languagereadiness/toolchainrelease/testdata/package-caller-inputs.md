# Package input observations

The five `package-caller-*inputs*.json.gz` files are compressed original command
outputs from clean compiler source `92c1b7f860bd7d72fe6a06c7bd5d6f727f43242b`,
built with Go 1.27.2 on macOS arm64. That development source still reports
`0.6.17-dev`; the public 0.6.17 binary predates the input-only command option.
The native release profile separately runs the current candidate on each target.

All use `examples/package-caller-construction/gooo.workspace.json` and its
unchanged construction, input and evaluation files. No model is loaded.

| Observation | Source receipt | Supplied file | Attempt budget |
| --- | --- | --- | --- |
| `partial-inputs` | fresh construction | `inputs.json` | 5 |
| `inputs` | fresh construction | `inputs.json` | 6 |
| `inputs-replay` | existing `package-caller-construct.json.gz` | `inputs.json` | saved |
| `inputs-again` | `inputs` above | `inputs.json` | saved |
| `inputs-evaluate` | `inputs-again` above | `evaluation-cases.json` | saved |

Fresh input observations use runtime v1. Saved execution uses its owned executor
and runtime v2. Both execute twice and keep zero inference. Input-only receipts
have no expected/pass labels and retain zero finite evaluation counts. The
final labelled evaluation has four supplied expected values. Original timestamps,
process timings and construction history are retained; gzip uses no filename or
timestamp header. Negative tests mutate copies of these records in memory.
