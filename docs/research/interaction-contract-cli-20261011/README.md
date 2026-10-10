# Interaction model through the native compiler

Observed once on 2026-10-11 KST, after freezing the source, inputs, model and
protocol. The compiler source is `73024efdd8f5bb877071460011bfb2137420e54a`,
Go 1.27.2, SDK 0.2.39. This local development observation predates publication
of the compiler integration. It does not repeat the SDK's original model study.

The original public interaction model proposes branch layout and comparison
operand order in a small absolute-value function. Gooo checks seven authored
outputs and three authored intermediate conditions. A fixed caller adds the
original input to the selected function's result. Some evaluation values overlap
the authored goals; these are integration cases, not held-out accuracy evidence.

## What happened

| Operation | Observed result |
| --- | --- |
| Input inspection | 2 × 528 source cells, 7 × 32 output cells, 3 × 32 condition cells; zero predictions or candidate tests |
| Deterministic construction | 2 candidates, all 7 outputs and 3 conditions satisfied |
| Model construction | First proposal failed; second candidate satisfied all authored goals; one initial prediction |
| Native construction | One initial model prediction, 11/11 caller outputs satisfied |
| Saved construction replay | Zero new predictions, 11/11 outputs satisfied |
| Saved-outcome comparison | All 11 input groups unchanged and still matching |

Candidate attempts did not improve in this example. Compiler-reported code
construction took **1.335625ms without the model and 6.862250ms with it**. Model
loading accounted for 5.003375ms of the latter. Prediction took 137.667µs in the
body-generation command and 113.083µs in native construction. These are two
individual observations from different commands, not a latency distribution.

| Whole command | Wall time | CPU user + system | Max resident memory reported by time |
| --- | ---: | ---: | ---: |
| Deterministic body generation | below the timer's 0.01s resolution | below timer resolution | 20.55MiB |
| Model body generation | 0.01s | 0.01s | 22.27MiB |
| Native construction with model | 0.95s | 0.51s | 83.09MiB |
| Saved native replay | 0.41s | 0.28s | 82.45MiB |

For native construction, CPU time divided by wall time corresponds to 53.68% of
one CPU core on average; replay gives 68.29%. This includes compilation and child
process work. It measures neither the laptop's overall utilization nor the model's
isolated CPU increase. The short body-generation commands are too brief for a
useful CPU percentage at this timer resolution. Raw observations also include the
initial inspection command, which took 0.54s in this fixed order.

The order was fixed, after regression tests, with uncontrolled compiler and OS
caches. Memory is the maximum resident figure from `/usr/bin/time -lp`, not the
sum of simultaneously resident processes. No causal speedup or production cost
saving follows from this one execution. The 76,136-byte model weight array is
only one part of command memory. Customer savings and external adoption remain null.

## Inspect and verify

The six compressed JSON files preserve complete original stdout, including the
selected Go body and exact int64 observations. Timer output is retained separately.
No model weights or executable binaries are duplicated here. The public model's
SHA256 is `4b61fe8d84df77c8dfac6a880eedcddfd8f5f903fa538d1e80532dd77f74b1ac`.

```sh
sh docs/research/interaction-contract-cli-20261011/verify.sh
```

This reads saved files only. It checks checksums, channel widths and hashes,
matching inspection/ranking inputs, prediction counts, finite outcomes and the
unchanged saved replay. It does not run the compiler, execute generated code,
download weights, fit a model or repeat these six commands.

The first local CLI adapter test reused an older three-choice example. Its
arithmetic choice correctly declined the new representation, so the test's
expectation of one prediction failed. Coverage now checks that deterministic
continuation separately from a supported two-choice source. A later independent
score test initially omitted the SDK's condition-feature-version method; that
test adapter was corrected. Focused race tests and both affected packages' full
unit suites passed. The read-only audit was corrected to use the existing raw
case digest and nested `search.new_local_model_predictions` JSON field. Original
compiler/model observations were never rerun to make these checks pass.

See [the source and commands](../../interaction-contract-model.md) for the input
format and current limits. An unsupported representation remains an explicit
deterministic fallback; it is not counted as a successful model prediction.
