# Package interface fixture

`package-interface.json` is the exact JSON output of compiler source
`8f4c96386a50fb7d79072bb5a743ef276b658214`, built with Go 1.27.2:

```sh
gooo package interface --json examples/package-optional-record-flow/gooo.workspace.json
```

The two source files declare one three-field optional record, its owning domain
activity and the importing app activity. The release profile checks these
declarations and original source digests, repeated byte equality and plain output.
This fixture describes source structure; it contains no activity execution or
compatibility verdict. The optional native test runs the supplied packaged CLI
through the same profile used by all release platforms.
