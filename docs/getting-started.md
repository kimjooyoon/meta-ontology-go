# Your first Gooo program

Gooo declares a program's intent, permitted construction choices and finite
examples, then generates Go and records what ran. This walkthrough builds a
diagnostic tool and reuses it on new inputs. It needs no model or API key.

## Install the matching compiler

Use **Go 1.27.2** for native execution, matching your Gooo executable's operating
system and architecture. Check with `go version`. The compiler below is the
experimental [v0.6.24-dev release](https://github.com/kimjooyoon/meta-ontology-go/releases/tag/v0.6.24-dev).
The release is pinned so these instructions do not silently select later changes.

On macOS or Linux, start in a new working directory and install locally:

```sh
GOBIN="$PWD/.gooo-bin" go install github.com/kimjooyoon/meta-ontology-go/cmd/gooo@v0.6.24-dev
export PATH="$PWD/.gooo-bin:$PATH"
gooo version --build --json
```

The version output includes `version: "0.6.24-dev"` and module version
`v0.6.24-dev`. Keep this compiler with your experiment for later replay.
The `export` applies to this shell; a new shell needs the same absolute directory
on its `PATH`.

Alternatively, download and extract the matching archive, then add its directory
to `PATH` (or invoke the extracted `gooo` / `gooo.exe` by its full path):

| Platform | v0.6.24-dev archive |
| --- | --- |
| macOS Apple Silicon | [darwin-arm64](https://github.com/kimjooyoon/meta-ontology-go/releases/download/v0.6.24-dev/gooo-darwin-arm64.tar.gz) |
| macOS Intel | [darwin-amd64](https://github.com/kimjooyoon/meta-ontology-go/releases/download/v0.6.24-dev/gooo-darwin-amd64.tar.gz) |
| Linux x86-64 | [linux-amd64](https://github.com/kimjooyoon/meta-ontology-go/releases/download/v0.6.24-dev/gooo-linux-amd64.tar.gz) |
| Windows x86-64 | [windows-amd64](https://github.com/kimjooyoon/meta-ontology-go/releases/download/v0.6.24-dev/gooo-windows-amd64.zip) |

Compare the archive's SHA-256 with the published
[SHA256SUMS](https://github.com/kimjooyoon/meta-ontology-go/releases/download/v0.6.24-dev/SHA256SUMS).
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

## Scope and feedback

The language and receipt formats are experimental. These finite examples do not
prove all-input correctness. The starter returns a suggested operation as data;
your application decides what to do with it. Package distribution uses source
files and a workspace manifest; there is no Gooo package registry in this flow.

Try a separate source-porting use case in [Gooo Go Ports](https://github.com/kimjooyoon/gooo-go-ports).
It compares eight HTTP/Unicode functions against pinned Go originals and retains
[results and limits](https://github.com/kimjooyoon/gooo-go-ports#검증-범위).

If this walkthrough fails, [open a compiler issue](https://github.com/kimjooyoon/meta-ontology-go/issues/new)
with your operating system/architecture, `go version`, `gooo version --build --json`,
the exact command and diagnostic, and the smallest relevant source/input example.
Remove private input data first. For an idea or question, use
[Discussions](https://github.com/kimjooyoon/meta-ontology-go/discussions).
Implementation contributions follow [CONTRIBUTING.md](../CONTRIBUTING.md).
