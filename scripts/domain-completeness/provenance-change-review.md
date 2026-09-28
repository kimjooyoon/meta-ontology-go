# Provenance change review

A provenance-bound change proposal can be reviewed without applying it. `ReviewProvenanceChange` requires proposal, reviewer, approval-evidence, and application-boundary digests before returning `ACCEPTED_FOR_APPLICATION`.

That state means only that a separate application step may be entered. It does not mutate a catalog, execute generated code, or grant authorization. Missing evidence remains `PENDING` and cannot be promoted by cache presence or an inferred score.
