# Ask Gooo which bytes a digest field describes

Two agent readers compared valid hashes from different representations: a raw
Gooo file against a workspace JSON-string digest, and a published release file
against the report's own typed-value digest. This example makes those known
field meanings executable Gooo data. Its decisions live in `main.gooo`; the
existing package runner supplies native execution and saved replay.

From the repository root, using Go 1.27.3:

```sh
go run ./cmd/gooo package execute --json \
  --cases examples/source-digest-guide/cases.json \
  examples/source-digest-guide/gooo.workspace.json \
  > /tmp/gooo-source-digest-guide.json
```

Each query supplies the exact owning command/document and full field path.
The result echoes both and returns `known`, `byte_domain`, `source_scope` and
`hash_recipe`. For example:

```json
{"owner":"package execute","path":"result.program.workspace_image.packages[].sources[].source_digest"}
```

This selects Go JSON-string encoding of the complete workspace source.
`result.program.activity_origins[].input_source.source_sha256` selects raw UTF-8
bytes of the complete original file, including when its span selects a smaller
region. The generation report selects its own input checkpoint, which can differ
after an earlier body fill. The release report's typed-value digest and its
published file checksum select separate representations.

The catalogue covers eight known owner/path pairs. A leaf name alone, an unknown
owner, or a changed path returns `known=false`, `UNKNOWN`, `UNRESOLVED` and
`READ_PRODUCER_CONTRACT`. Matching is exact; the program preserves spelling and
does not trim or guess aliases. The `hash_recipe` strings explain the producer's
operation and are not shell commands or a general expression language.

The twelve cases include all eight known pairs and four unresolved queries:
a leaf-only path, a path under the wrong command, a trailing space in the owner,
and an undeclared Korean owner spelling. The source itself contains Korean text,
quotes and line breaks, so its original-file and Go JSON-string hashes differ.

This is an advisory catalogue for the documented producer contracts. It does
not read a receipt or file, calculate a hash, authenticate an owner, or determine
whether two files are equal. A caller still selects the intended source entry
and uses the existing receipt reader or replay validation. The normal execution
path performs no additional work for this example. Its finite query cases check
the guide's behavior; existing producer regressions independently check the
workspace encodings and wrong-representation refusal.

The ordinary package receipt retains `Explain`'s stable activity ID and the
original `main.gooo` file digest. The output's `owner` is the caller's spelling;
the receipt's source origins identify the program that supplied the explanation.
The fixed Gooo evaluator, native execution and saved replay use the same twelve
query inputs, rather than three separate populations.

The [digest table](../../docs/getting-started.md#read-a-digest-with-its-full-field-path)
links these paths to their producer contexts. The wider inventory remains
[issue #1728](https://github.com/kimjooyoon/meta-ontology-go/issues/1728).
