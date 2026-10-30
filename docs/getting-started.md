# Your first Gooo program

[Try the score-and-label tutorial](https://kimjooyoon.github.io/gooo-go-ports/tutorial.html)
without installation or a model. Change two inputs and inspect four returned
fields, then download the source and inputs for local editing and execution.
It runs a fixed, precompiled Gooo → Go → WebAssembly program; it does not compile
arbitrary source. The [library explorer](https://kimjooyoon.github.io/gooo-go-ports/)
provides separate examples of selected Go function ports.

Gooo declares a program's intent, permitted construction choices and finite
examples, then generates Go and records what ran. This walkthrough builds a
diagnostic tool and reuses it on new inputs. It needs no model or API key.

## Install the matching compiler

Use **Go 1.27.3** for native execution, matching your Gooo executable's operating
system and architecture. Check with `go version`. The compiler below is the
experimental [v0.6.39-dev release](https://github.com/kimjooyoon/meta-ontology-go/releases/tag/v0.6.39-dev).
The release is pinned so these instructions do not silently select later changes.

On macOS or Linux, start in a new working directory and install locally:

```sh
GOBIN="$PWD/.gooo-bin" GOWORK=off GOTOOLCHAIN=go1.27.3 \
  go install github.com/kimjooyoon/meta-ontology-go/cmd/gooo@v0.6.39-dev
export PATH="$PWD/.gooo-bin:$PATH"
gooo version --build --json
```

The version output includes `version: "0.6.39-dev"`. Module and VCS fields depend
on the installation method; the release manifest binds the public archives to
source `4bfb81640fcf485a022abc3b0484b2fd7bcb9f65`, which pins SDK `v0.2.51-experimental`.
Native archives report `source_status: "CLEAN_VCS"` and that source commit.
A versioned module installation reports `module.version` and `module.sum`; its
VCS fields remain `UNOBSERVED_VCS`. Archive module versions can be pre-tag
pseudoversions. Record the language `version` and the reported identity fields
with the installation method. Keep this compiler with your experiment for replay.
The `export` applies to this shell; a new shell needs the same absolute directory
on its `PATH`.

Alternatively, download and extract the matching archive, then add its directory
to `PATH` (or invoke the extracted `gooo` / `gooo.exe` by its full path):

| Platform | v0.6.39-dev archive |
| --- | --- |
| macOS Apple Silicon | [darwin-arm64](https://github.com/kimjooyoon/meta-ontology-go/releases/download/v0.6.39-dev/gooo-darwin-arm64.tar.gz) |
| macOS Intel | [darwin-amd64](https://github.com/kimjooyoon/meta-ontology-go/releases/download/v0.6.39-dev/gooo-darwin-amd64.tar.gz) |
| Linux x86-64 | [linux-amd64](https://github.com/kimjooyoon/meta-ontology-go/releases/download/v0.6.39-dev/gooo-linux-amd64.tar.gz) |
| Windows x86-64 | [windows-amd64](https://github.com/kimjooyoon/meta-ontology-go/releases/download/v0.6.39-dev/gooo-windows-amd64.zip) |

Compare the archive's SHA-256 with the published
[SHA256SUMS](https://github.com/kimjooyoon/meta-ontology-go/releases/download/v0.6.39-dev/SHA256SUMS).
These archives contain the compiler; native execution still needs Go 1.27.3.
When Go is elsewhere, add `--go /path/to/go1.27.3` to each execute/replay command.
On Windows, run the commands below in **Command Prompt (`cmd.exe`)**, using
`gooo.exe`. Windows PowerShell 5.1 writes UTF-16LE with `>`, which makes the saved
receipt unreadable; use Command Prompt to preserve the program's UTF-8 output.

## Build once, then run new inputs

Choose a directory that does not already exist:

```sh
gooo init --template diagnostic my-diagnostic
cd my-diagnostic
gooo package execute --json --cases cases.json gooo.workspace.json > execution.json
gooo package replay --receipt execution.json --inputs inputs.json gooo.workspace.json
```

The last command prints exactly:

```text
"partial: missing branch result [repair-and-replay]"
"unobserved: Add expected observations. [add-examples]"
"complete: All observed fields matched. [accept]"
```

`diagnostics.gooo` owns the conditions and permitted field choices. Its five
construction examples help select a body. `app.gooo` imports that activity and
formats its result with this Gooo declaration:

```gooo
activity Main(Diagnostic) -> Text computes `return input.code + ": " + input.message + " [" + input.action + "]"`
```

The execute command compiles generated Go and checks eight named outputs across
four input rows. `execution.json` retains the selected program and observations.
Replay uses the saved choices on three additional input rows without a new search
or model call. These input-only rows have no expected answers and earn no correctness score.

The generated `inputs.json` has one object per input row. Its first row has this
shape (the generated file contains two more rows):

```json
{
  "schema": "gooo/body-composition-inputs/v1",
  "inputs": [{
    "diagnostic/diagnostics:Diagnose.input0": 1,
    "diagnostic/diagnostics:Diagnose.input1": 3,
    "diagnostic/diagnostics:Diagnose.input2": 1,
    "diagnostic/diagnostics:Diagnose.input3": "missing branch result"
  }]
}
```

For activities with multiple inputs, read each key as
`<package path>:<activity>.<input port>`. The diagnostic starter uses module
`diagnostic` even when its directory is named
`my-diagnostic`; `gooo.workspace.json` supplies the package path and
`diagnostics.gooo` declares the activity's positional inputs. If you pass
`--module` to `gooo init`, use that module in the keys. An activity with one
input uses the whole `<package path>:<activity>` key, whether its value is a
scalar or a record. The [scope-fragments input](../examples/scope-fragments/inputs.json)
shows a record value.

## Change something useful

Open `inputs.json`. The four `Diagnose` input keys are matched outputs (`input0`),
total outputs (`input1`), rejected candidates (`input2`) and detail (`input3`).
Change the first row's detail to `my missing result` and repeat the replay command.
Its first value becomes `"partial: my missing result [repair-and-replay]"`.

Change conditions or permitted choices in `diagnostics.gooo` to change the tool's
behavior. Update its selection examples and `cases.json` expectations deliberately,
then repeat `package execute` to create a fresh receipt. A source edit invalidates
the old receipt; keep the old source and receipt together if you need both versions.
The [starter guide](language/project-starters.md) also explains imported library projects.

## Use results from an agent or CI

Use `--json` when another program consumes the result. Keep stdout, stderr and
the process exit code separately. Missing or malformed JSON does not satisfy an
acceptance rule. The command's exit code reports whether the
operation completed; the receipt records what was observed:

| Result | Meaning and next action |
| --- | --- |
| `decision: "PASS"`, exit 0 | All supplied expected observations matched. Read `result.runtime.finite_passed` and `finite_total` for this finite scope. |
| `decision: "PROGRESS"`, exit 0 | The operation completed and some supplied expectations differed. Inspect the mismatched fields; a CI acceptance rule should reject this result when it requires all supplied cases to match. |
| `decision: "OBSERVED"`, exit 0 | Input-only observation without expected answers. Inspect the actual values and per-row fault/blocked details; there is no correctness percentage. |
| `decision: "FAIL_CLOSED"`, exit 1 | The operation failed. Read `error` and any retained execution details before changing the source or environment. |
| Exit 2 | The command arguments were rejected. Usage is written to stderr; stdout need not contain a JSON receipt. |

`finite_total` counts entries in each row's `expected` object; `finite_passed`
counts entries whose complete JSON values match. A whole-record expectation is
one comparison even when the record contains several fields. Keep input-row
counts and declared field-value counts separate when reporting a result.

For the pinned walkthrough's `--cases` command, a shell CI step can use its
existing JSON parser to require `PASS`, a positive `finite_total`, and equal
`finite_passed` / `finite_total`. For example, in a POSIX shell with **jq** available:

```sh
set -eu
gooo package execute --json --cases cases.json gooo.workspace.json > execution.json
jq -e '.schema == "gooo/workspace-body-execution-receipt/v1" and
       .decision == "PASS" and .error == null and
       .result.runtime.finite_total > 0 and
       .result.runtime.finite_passed == .result.runtime.finite_total' execution.json
```

Run these as separate CI steps or stop the shell on command failure. The JSON
assertion is additional to the execute command's exit status. An input-only
`OBSERVED` receipt does not satisfy this case-matching assertion. Choose expected
outputs independently of the implementation you are checking.

Keep the compiler version, source, workspace, inputs/cases and receipt together.
After editing `.gooo`, run `package execute` again to select and save the new
program. `package replay` checks the saved choices against the source and really
executes the new inputs; it does not return saved outputs. A changed source can
therefore reject an older receipt. Keep the old source with it if both versions
are needed.

Consumer CI checks the application's own source, contracts and cases with its
pinned compiler. The compiler repository's internal six-check pipeline and
experimental observation workflows are not installation requirements for this
walkthrough. No model setup is needed. A later compiler upgrade is a separate
compatibility change with its own cases.

### Read a digest with its full field path

Gooo exposes several source and artifact representations. Read each field in the
command result, plan or release record that contains it; the same digest prefix
does not imply the same input bytes:

| Field and owning record | What it describes |
| --- | --- |
| `manifest_digest`, `inputs_digest`, `cases_digest` | The bytes of the corresponding supplied file, when that input mode is present. |
| `gooo package resolve --json`: `result.image.packages[].sources[].source_digest` | SHA-256 of `json.Marshal(source.Content)`, where `Content` is the source string loaded from the file. |
| `result.program.workspace_image.packages[].sources[].source_digest` | The stored source string serialized by Go `json.Marshal`, including JSON quoting and escaping. This is different from hashing the raw `.gooo` file. |
| `result.program.activity_origins[].input_source.source_sha256` | Raw UTF-8 bytes of the complete workspace source file named by `package_path` and `filename`. The accompanying byte offsets select a region within that file; the digest does not hash only that region. |
| `result.program.workspace_image.packages[].sources[].semantic_digest` | The compiler's semantic IR digest for that source entry. |
| `result.composition.steps[].generation.report.source_digest` | SHA-256 of the raw Gooo source bytes passed to that generation step. Earlier fills or selected source checkpoints can change these bytes between steps. A direct generation report uses `report.source_digest` for its own input bytes. |
| `result.composition.steps[].generation.report.body_fill.original_source_digest` or `.body_search.original_source_digest` | SHA-256 of the raw Gooo bytes supplied to that fill or search operation, before its own choice is applied. “Original” is relative to that operation. |
| `result.execution_source_sha256` | The combined, lowered Gooo source after applying body fills, passed to composition execution. If fills changed it, it differs from the initial `result.program.lowered_gooo_source`. |
| `ir-search` plan input: `external_training_feedback.source_digest` | SHA-256 of the raw Gooo source bytes supplied to that search invocation. In direct file mode these are the file bytes; a composed caller may supply a selected source checkpoint. |
| `release-manifest.json`: `version_source_digest` | Raw bytes of `internal/meta/releaseidentity/version.gooo`. The publication contract's `identity_source` also names the path and stable activity ID. |
| `release-manifest.json`: `release_report_digest` | The report's own `report_digest`: SHA-256 of Go's typed Report JSON with its `report_digest` field set to an empty string. |
| `release-manifest.json`: `assets[name="release-report.json"].digest` | Raw bytes of the complete published report file, including formatting and its populated `report_digest`. The corresponding `SHA256SUMS` entry uses these same bytes. |
| `release-manifest.json`: `publication_contract_digest` | Raw bytes of `contract.json` retained by the original publication workflow artifact. |

For a downloaded release file, compare its raw-file checksum with the matching
`assets[].digest` or `SHA256SUMS` entry. The existing
[report validator](../internal/meta/languagereadiness/toolchainrelease/validate.go)
recomputes the separate typed-report digest. Both values were checked for the
[0.6.39 release](releases/0.6.39-dev.md); a wrong-representation comparison refused
in the private reader before its field interpretation was corrected.

Match the source entry by package path and filename. Use the existing replay
command to validate a saved program against its current sources; do not replace
that check with equality between unrelated digest fields. When reporting a
mismatch, retain the compiler version, full JSON path and the bytes or
representation you compared. The source string's JSON digest is not a raw file
checksum.

For a source string containing Korean text, quotes and newlines, the two
computations are `sha256(json.Marshal(sourceString))` and
`sha256([]byte(sourceString))`. They produce different values. Passing the
raw-byte value as a package image's `source_digest` is rejected as stale type
evidence; it does not prove that the file changed. The normal package path
still computes only its existing image digest.

The runnable [Gooo digest guide](../examples/source-digest-guide/README.md) accepts
an owning command/document and an exact field path, then explains eight known
byte domains. Unknown combinations stay unresolved. It uses the existing package
runner and adds no hashing or receipt authority to ordinary execution.

### Read values through the runtime trace

For application results, read `result.runtime.traces[].deliveries[]`: each row
keeps the activity ID, `actual` value and any `fault` or `blocked_by` condition.
Use `case_index` to associate it with the supplied input row. Expected values
belong to `--cases`; input-only observation keeps the finite matching counts at zero.

The saved composition and the executable observation driver have separate origins:

| Receipt field | Meaning |
| --- | --- |
| `result.composition.driver_source` and `driver_sha256` | The saved composition's base driver and its source digest. |
| `result.runtime.observed_driver_sha256` | The driver actually derived for observation, when its source differs from the base driver. |
| `result.runtime.runs[].stdout_sha256` | The exact bytes emitted by that native process, including its observation envelope and encoding. |

For example, the arithmetic observer emits
`gooo/arithmetic-outcome-observation/v1` with `outputs`, `faults`, `blocked` and
`calls`, even when that run has zero faults. The outer runtime receipt can still
use `gooo/body-composition-runtime/v1` in this case. Its version alone therefore
does not identify the native stdout shape. A consumer that assumes the base
driver's plain output matrix can reject a successful observation.

The runtime trace already exposes the decoded values and conditions. Official
replay checks the saved program and executes the new inputs through its own
producer contract. If an independent audit needs native byte hashes, retain the
exact producer revision and observation mode; re-encoding just the `actual`
values describes different bytes. The [arithmetic observation notes](research/native-arithmetic-outcomes-20261009.md)
describe this particular observation path.

## Variables, conditions and source origins

The development source includes a small [ordinary program](../examples/everyday-program/README.md).
It adds 11 to an input and doubles the result when the Boolean input is true:

```gooo
activity Adjust(Integer, Boolean) -> Integer id "urn:gooo:everyday:adjust" computes `let value=input0+11
if input1 {value=value*2}
return value`
```

`let` introduces a local, and `value=value*2` assigns its next value. The activity
ID keeps the declaration's identity as its display name changes. From that example's
directory, `package execute` runs its declared cases; `body-origins --activity Adjust`
reads possible input and control dependencies from the source.

The saved development observation passed three finite outputs with zero model
calls and resolved 16 origin nodes. Its source manifest and build metadata bind
producer `10067b26de6d3fa75040173a610fdc522b5da62a`, Go 1.27.2 and SDK
`v0.2.51-experimental`. Those observations retain that producer identity; the public
archive installed above has its separately pinned source. The example explains
both the ordinary program and the origin report's scope.

## Edit Gooo in VS Code

The existing [Gooo extension](https://github.com/kimjooyoon/gooo-ecosystem-workbench/tree/vscode-v0.2.5/editors/vscode)
provides a `source.gooo` TextMate grammar, snippets and compiler commands.
Its [0.2.5 release](https://github.com/kimjooyoon/gooo-ecosystem-workbench/releases/tag/vscode-v0.2.5)
contains a VSIX package for **Extensions: Install from VSIX...**.
Set **`gooo.compilerPath`** to the absolute path of the compiler installed above,
such as the `gooo` executable in your `.gooo-bin` directory.

Open the generated project folder and use **Gooo: Check Current File** and
**Gooo: Format Document**. The extension also connects to the compiler's
experimental language server; editor and compiler compatibility remains under
development. Leave the optional provider settings empty for deterministic construction.

The [language product priorities](language/language-product-priorities.ko.md)
describe the planned improvements to everyday programs, reuse and editor support.

## Scope and feedback

The language and receipt formats are experimental. These finite examples do not
prove all-input correctness. The starter returns a suggested operation as data;
your application decides what to do with it. Package distribution uses source
files and a workspace manifest; there is no Gooo package registry in this flow.

Try a separate source-porting use case in [Gooo Go Ports](https://github.com/kimjooyoon/gooo-go-ports).
It compares selected Go standard-library functions against pinned Go originals.
Its [results and limits](https://github.com/kimjooyoon/gooo-go-ports/blob/75d0655101d1468f9a94eb10571d34378b98293b/README.md)
distinguish finite native cases, Unicode enumeration and browser checks.

If this walkthrough fails, [open a compiler issue](https://github.com/kimjooyoon/meta-ontology-go/issues/new)
with your operating system/architecture, `go version`, `gooo version --build --json`,
the exact command and diagnostic, and the smallest relevant source/input example.
Remove private input data first. For an idea or question, use
[Discussions](https://github.com/kimjooyoon/meta-ontology-go/discussions).
Implementation contributions follow [CONTRIBUTING.md](../CONTRIBUTING.md).
