# Capability-scoped expansion

`main.gooo` is the one semantic source used for every case. Its `computes`
value programs define policy, capability kind/operation/target, prior claim
state, evidence class, and eight cases. Activity names and comments are not
authority.

The CI provider actually reads one pinned file from a temporary sandbox and
one deterministic logical input. Environment and network are not contacted, so
they are `UNKNOWN` or `HISTORICAL_FIXTURE`, not current evidence. The same
provider observes denied write, mutation, and promotion requests through
sandbox before/after snapshots.

The independent consumer parses and lowers the raw source again, consumes raw
provider observations, and does not import the producer. CI also runs a
semantic-policy intervention and a comment-only intervention. This is a
bounded experiment, not a general macro sandbox.
