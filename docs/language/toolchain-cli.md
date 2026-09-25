# Deterministic toolchain CLI

## Boundary

`TOOLCHAIN-CLI` credits the existing `cmd/gooo` executable only after CI builds
that exact head and a separate metaprogram observes its process contract. The
implementation does not add another router, shell, interpreter, plugin loader,
network client, or repository writer.

The observer accepts one absolute binary outside the repository and maps a
versioned registry to twelve compiled-in argument plans. Registry text cannot
select an arbitrary executable or arbitrary arguments. Each plan runs twice
with a five-second deadline and a 64 KiB output bound.

## Useful properties

- Text and JSON identity are both executable contracts.
- Syntax, semantic, and bidirectional language operations cross the real CLI.
- Exit code, stdout, stderr, and repository tree state are independently
  measured rather than summarized as one green result.
- Each invocation records a positive integer `peak_rss_kib` observation from the
  runner. This value is `RUNNER_SCOPED_NONDETERMINISTIC`: it is visible evidence,
  not replay authority and not a memory or performance improvement claim.
- Unknown observations lower resolution and cannot become readiness credit.

The [gomacro project](https://github.com/cosmos72/gomacro) is used only as a
structural reference for separating a command front door from staged language
evaluation. This project neither imports gomacro nor claims its REPL,
interpreter, macro, or dynamic package capabilities.

No novelty claim is made. The positive element is the combination of a real
binary boundary, fixed meta operations, deterministic replay, and zero-write
evidence in one readiness receipt.
