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

Use **Go 1.27.2** for native execution, matching your Gooo executable's operating
system and architecture. Check with `go version`. The compiler below is the
experimental [v0.6.34-dev release](https://github.com/kimjooyoon/meta-ontology-go/releases/tag/v0.6.34-dev).
The release is pinned so these instructions do not silently select later changes.

On macOS or Linux, start in a new working directory and install locally:

```sh
GOBIN="$PWD/.gooo-bin" GOWORK=off GOTOOLCHAIN=go1.27.2 \
  go install github.com/kimjooyoon/meta-ontology-go/cmd/gooo@v0.6.34-dev
export PATH="$PWD/.gooo-bin:$PATH"
gooo version --build --json
```

The version output includes `version: "0.6.34-dev"`. Module and VCS fields depend
on the installation method; the release manifest binds the public archives to
source `b266c0e2cf6878385b59788945861a03e33f766b`, which pins SDK `v0.2.44-experimental`.
Keep this compiler with your experiment for later replay.
The `export` applies to this shell; a new shell needs the same absolute directory
on its `PATH`.

Alternatively, download and extract the matching archive, then add its directory
to `PATH` (or invoke the extracted `gooo` / `gooo.exe` by its full path):

| Platform | v0.6.34-dev archive |
| --- | --- |
| macOS Apple Silicon | [darwin-arm64](https://github.com/kimjooyoon/meta-ontology-go/releases/download/v0.6.34-dev/gooo-darwin-arm64.tar.gz) |
| macOS Intel | [darwin-amd64](https://github.com/kimjooyoon/meta-ontology-go/releases/download/v0.6.34-dev/gooo-darwin-amd64.tar.gz) |
| Linux x86-64 | [linux-amd64](https://github.com/kimjooyoon/meta-ontology-go/releases/download/v0.6.34-dev/gooo-linux-amd64.tar.gz) |
| Windows x86-64 | [windows-amd64](https://github.com/kimjooyoon/meta-ontology-go/releases/download/v0.6.34-dev/gooo-windows-amd64.zip) |

Compare the archive's SHA-256 with the published
[SHA256SUMS](https://github.com/kimjooyoon/meta-ontology-go/releases/download/v0.6.34-dev/SHA256SUMS).
These archives contain the compiler; native execution still needs Go 1.27.2.
When Go is elsewhere, add `--go /path/to/go1.27.2` to each execute/replay command.
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

The existing [Gooo extension](https://github.com/kimjooyoon/gooo-ecosystem-workbench/tree/vscode-v0.2.4/editors/vscode)
provides a `source.gooo` TextMate grammar, snippets and compiler commands.
Its [0.2.4 release](https://github.com/kimjooyoon/gooo-ecosystem-workbench/releases/tag/vscode-v0.2.4)
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
