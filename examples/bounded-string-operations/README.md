# Bounded string operations written in Gooo

The library implements substring search, non-overlapping occurrence counting and
UTF-8 rune counting in `.gooo`. An importing Gooo app returns all three results.
Its inputs are valid UTF-8 Text of at most 1,024 bytes.

`Index` returns a byte offset or -1. `Count` advances past each match. An empty
needle returns the rune count plus one, including 1 for empty Text. `MatchAt`
compares bytes in a separate bounded helper; both callers have their own counted
loops. `RuneCount` counts UTF-8 leading bytes within the valid Text contract.

From this directory, with public Gooo 0.6.32 or source containing counted-loop and
Text-index support, and Go 1.27.2 available:

```sh
gooo package execute --json --cases cases.json gooo.workspace.json > execution.json
gooo package replay --json --receipt execution.json --cases cases.json gooo.workspace.json
```

The committed cases and native CLI regression compare 20 cases with Go `strings.Index`,
`strings.Count` and `utf8.RuneCountInString`. Native execution and saved replay
each passed 20/20 with zero model calls. Cases cover empty needles, overlapping
matches, absence, NUL, newlines, Korean text, emoji and the 1,024-byte boundary.
These observations describe the supplied cases and the valid UTF-8 Text contract.

The local preparation used the clean compiler checkout `c56007f5` with an added
test-file overlay. Its native record reports `UNBOUND_LOCAL_SOURCE` for the
producer revision. The committed regression checks the actual example files;
CI provides the results for the submitted source revision.
