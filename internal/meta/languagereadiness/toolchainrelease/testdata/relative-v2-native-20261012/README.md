# Retained V2 native API observation

These are exact gzip-compressed JSON bytes from one fresh zero-weight V2
release-profile observation on darwin/arm64, Go1.27.2, source
`7f57c454603d40465ac3e3764451fe26f5143062` and public SDK0.2.42. The binary's
VCS state was clean and no local module replacement was used.

Preflight has zero model/candidate calls. Construction uses one zero-weight
ranking and four finite source candidates, then native caller1/1 and evaluation8/8.
The zero model file was deleted before a single saved native replay, which passed
8/8 with zero new model calls and unchanged selected program identity.

The initial smoke reader stopped after successful construction because it compared
an unprefixed condition hash with a prefixed hash. The observation was retained;
construction was not repeated. Reader corrections used existing records. Runtime
schema is v2 for the consumed caller, v1 for initial evaluation and v2 for saved
evaluation. The reader now binds those explicit scopes and all outcome counts.

This is an API/release regression, separate from the trained model's quality study.
It changes no original fit, model-quality result or performance measurement.
The fixture reader runs no inference or native executable. It reconstructs the
bounded source candidates to check original local outputs/Boolean history.

Decompressed file SHA256:

```text
f3f7f1bdea1c76416f6ac72a1800c3a0299eb3ae44929d17d6effbb723479551  preflight.json
f8d883ace34e716c08700716ee4833145b416aef2274c94ce94eb2317c4da9aa  construct.json
677aee2681bff99bebe56ba708a2fe287f0cdc0507690264cb42c522a0d1066e  replay.json
```

These files retain their original producer identity. They do not replace the next
release's exact accepted-source four-platform native witness.
