# Deterministic Query Readiness Floor

The deterministic-query binding proves that `LANGUAGE-DETERMINISTIC-QUERY` remains satisfied once readiness is at least `16/24` and `6666` basis points. Those values are the versioned minimum achieved by that concept, not a frozen current-project value.

A later concept may increase the same `24`-obligation registry. The binding records the observed current values but accepts them only when concepts are `>=16`, completed obligations are `>=16`, total obligations remain exactly `24`, readiness is `>=6666`, and all query-specific coordinates remain exact.

This prevents a valid later improvement from invalidating predecessor evidence while still failing closed for a lower or unknown state.
