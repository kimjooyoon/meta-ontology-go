# Billing operation-manifest experiment

Two Gooo source files define one symbolic user operation:
`PayOrder(Order) -> Receipt`. The compiler emits a JSON operation manifest,
not Go source and not an executable business implementation.

The fixed experiment observes:

- one primary manifest and one byte-identical replay;
- three content-bound artifact digests;
- five ordered wall-time and peak-RSS samples;
- one unknown-emitter rejection and five evidence-integrity counterexamples;
- zero repository writes and zero mutation authority.

The report exposes 6 user, 12 tool-author, and 15 governor coordinates. A
passing report means only `MINIMAL_VALUE_OBSERVED`. It does not claim business
correctness, value-level computation, production readiness, performance beyond
the fixed runner samples, or general-purpose code generation.

This experiment is read-only evidence. Promotion into self-improvement input is
deliberately deferred until content integrity and impossible resource samples
are proven fail-closed by CI.
