# Language operation catalog observer

The command emits a deterministic receipt for the current catalog source.
Both the intentional `0/1` baseline and a later exact `1/1` extension are
valid observable states; their decision names and reader views differ.

```sh
go run ./scripts/language-operation-catalog \
  -source examples/language-operation-catalog/main.gooo \
  -head-sha "$HEAD_SHA" -output "$OUTPUT/receipt.json" -check
```
