# Filename classification assembled in Gooo

This example supplies `HasPrefix`, `HasSuffix`, `StripSuffix` and `ByteLength`
as ordinary Gooo activities. `Classify` uses them to classify a single filename,
remove a final `.gooo` suffix and report its UTF-8 byte length. A leading dot
marks the name hidden; matching is case-sensitive. Callers supply a filename
rather than a path. No filesystem access occurs.

Three declared choices connect the computed values to the output. The source
checks five construction examples and permits eight attempts. The separate
native suite contains twelve other filenames, including Korean, an emoji,
a combining accent, a repeated extension and a trailing space.

## Run from a checkout containing Text operations

```sh
go run ./cmd/gooo body-compose --entry Classify \
  --source examples/text-operations/source.gooo.fixture \
  --cases examples/text-operations/cases.json --out /tmp/gooo-text-fixed
```

Use a fresh output directory. Add `--model /path/to/model.json` to use a compatible
local record-choice model. The compiler measures the same finite cases in either
mode. The source, selected candidates, generated Go and runtime observations are
saved together. Reuse the saved selection with zero new model calls:

```sh
go run ./cmd/gooo body-compose \
  --source /tmp/gooo-text-fixed/original.gooo \
  --composition /tmp/gooo-text-fixed/composition.json \
  --cases examples/text-operations/cases.json
```

## Text semantics

`len(text)` and `text[low:high]` follow the
[Go length](https://go.dev/ref/spec#Length_and_capacity) and
[slice](https://go.dev/ref/spec#Slice_expressions) rules. Positions and lengths
count bytes: `ByteLength("가")` is 3 and `ByteLength("🙂")` is 4.
An omitted low offset is zero; an omitted high offset is the byte length.
`int64(len(text))` converts the length to Gooo's Integer representation. A local
initialized from runtime `len` retains the target's `int` type until converted.

Prefix and suffix functions check length before slicing. Their short-circuit
condition handles empty and shorter inputs. Slicing at a UTF-8 boundary preserves
text; arbitrary byte offsets can split a character. JSON replaces invalid UTF-8
bytes, so use complete prefix/suffix boundaries for text passed through that
format. There is no Unicode normalization or locale-dependent comparison.

An evaluated invalid slice fails candidate interpretation and panics in an
unguarded native program, following the target language. The evaluator reports
the invalid offsets instead of crashing the compiler. Three-index slices,
byte indexing and loops remain outside this body profile.

Source-declared pure functions keep their identities and call limits. Value
primitives add no synthetic activity edges. Direct record expressions retain
their length, conversion and slice-bound dependencies in the symbolic value
graph. The optional source-flow export also follows the pure helper calls,
including their argument bindings and conditional return values. Its 512-node
and 64-live-binding bounds apply across helper frames. Source spans retain the
called activity identity. See [helper value flow](../../docs/record-value-flow.md#follow-a-pure-gooo-helper).

The existing origin model feature vocabulary still compresses different
operators together. Following a helper in the graph does not by itself make
those distinctions available to the current trained model.

The tests compare the interpreter, generated native program and saved replay.
Their counts describe these finite examples and this filename policy.
