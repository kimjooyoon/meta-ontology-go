# Deterministic toolchain format/fix

## Boundary

`gooo format` emits canonical Gooo source or checks that a source is already
canonical. `gooo fix` emits a versioned `gooo/format-fix-plan/v1` value. It
does not accept `--write`, open a repository writer, or apply its own plan.

The plan binds source bytes, result bytes, semantic fingerprints, one bounded
full-file edit, an explicit decision, and zero direct writes. A separate
metaprogram consumes the same plan in memory, rebuilds the result, and accepts
only an explicit `FIXED_POINT`. Unknown source meaning, plan decisions,
resolutions, registry operations, or top-level decisions fail closed and lower
resolution.

## Executable use cases

The fixed corpus contains twelve cases:

- six positive paths: format text, format JSON, canonical check, changed-plan
  JSON, fixed-point JSON, and plan text
- six guardrails: missing input, malformed format input, malformed fix input,
  unknown format option, unknown fix option, and rejected write authority

Every process case runs twice against the exact CI-built `gooo` binary.
Repository trees are measured before and after all twenty-four invocations.
The metaprogram separately performs one in-memory application and observes two
fixed points with zero direct writes.

## Structural reference

The [gomacro project](https://github.com/cosmos72/gomacro) remains a structural
reference for keeping a command boundary separate from staged language
evaluation. Gooo does not import gomacro and makes no claim to its interpreter,
REPL, macro, or dynamic package capabilities.

No novelty claim is made. The useful property is the measured combination of
canonical output, a data-only repair plan, semantic equality, in-memory
application, fixed-point replay, and a denied write boundary.
