# Capability discovery to investment bridge

This example connects two read-only language surfaces:

1. A natural-language question is resolved against declared capabilities and produces query, catalog, declaration, and provenance evidence.
2. The observation is compared with an explicit domain scope and investment target.
3. The bridge preserves `UNKNOWN`, `DEFERRED`, and `INVESTIGATE` before it can emit `INVEST` or `MAINTAIN`.

The target ratio is policy, not a truth score. Meeting it does not prove correctness or completeness. Any later execution or authorization remains an explicit external boundary and is not created by this bridge.
