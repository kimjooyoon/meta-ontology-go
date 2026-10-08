# 0.6.16 candidate: native arithmetic observations

Clean source `58c0c6f328bba18523e164857bd725c6f32db50f` built and replayed the
macOS arm64 executable and archive byte for byte using Go 1.27.1. The actual
candidate passed existing profiles plus five native-arithmetic profiles:
partial construction, complete construction, saved construction replay, graph
execution and graph replay. The platform receipt and five raw outputs are in
`native-platform-observations.tar.gz`. The compiler archive itself is not copied
into this source directory. This is candidate evidence; publication is separate.

The new validators retain original source/cases, all assignments, failed native
operands and location, graph inputs/dependencies, earlier and independent values,
process completion, exact integers and zero new inference. Release-package race
checks passed in 3.626 seconds, including 26 caller and 12 graph changed-evidence
cases. The CLI version regressions and vet also passed.

The same candidate executable then ran the original two related programs. An
independent Go recount reads JSON numbers exactly, checks original source and
expectations and compares selected source across fixed/model ordering and replay.
`summary.json` and six compressed command outputs retain these observations:

| Run | Attempts | Preflight rejections | Native combinations | With native faults | Final cases |
| --- | ---: | ---: | ---: | ---: | ---: |
| Budget 5 | 5 | 2 | 3 | 1 | 1/4 |
| Budget 6 | 6 | 2 | 4 | 1 | 4/4 |
| Saved budget | 6 | 2 | 4 | 1 | 4/4 |
| Mixed fixed | 48 | 16 | 32 | 8 | 4/4 |
| Mixed own model | 6 | 2 | 4 | 1 | 4/4 |
| Saved mixed | 6 | 2 | 4 | 1 | 4/4 |

Saved counts describe history; replay executes it again with zero new inference.
Native faults are a subset of executed combinations. Each complete native
observation has two fresh executions. The model and deterministic runs selected
identical source; original failures and separate final expectations remain intact.

The unchanged graph QAT model made one prediction in 29,000 ns. Metadata SHA-256:
`3c68205a660695103712e2115ce90accd298b406dced8ad5ad5c9b84bab0a202`;
weights SHA-256: `76f68845a03ed8c8bc261a57c919e96dcdd352d3c35bcdaeb3d9b51a88975c6f`.
It retains 2,096 resident tensor bytes and orders record choices. Fill choices
follow source order. No new training or model download was performed.

The single model command took 1.50 seconds wall time, 0.74 seconds user CPU,
0.51 seconds system CPU and 87,539,712 bytes maximum RSS (about 83.5 MiB).
Native compilation and execution are included. These process measurements do
not sample whole-machine utilization or establish a controlled speed comparison.
The prediction interval is separate from whole-command time.

The publication workflow and contract were subsequently updated from their old
version literals to 0.6.16. The recorded candidate above predates that metadata
and documentation update; final PR CI checks the submitted source. All original
raw observations retain their own clean source revision.
