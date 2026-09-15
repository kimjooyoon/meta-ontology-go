# Action Runtime Conformance

This directory contains the metaprogram that interprets the authoritative
GitHub Actions workflow as data.

## Contract

The tool reads one workflow, binds it to an exact commit SHA, and emits a
deterministic JSON report. It does not edit the workflow or write generated
files into the repository.

The embedded policy records the first official action major that runs on
Node.js 24. Each policy rule carries its authoritative upstream evidence URL
and participates in the policy digest.

## Indicators

- foundation.catalog-coverage requires every observed first-party action to
  have a policy rule.
- foundation.exact-head-binding requires a lowercase 40-character commit
  identity.
- coherence.node24-runtime requires every observed action reference to meet
  its Node.js 24 minimum.
- coherence.action-input-schema rejects inputs absent from the catalogued
  action schema.
- regression.canonical-replay evaluates the projection twice and requires
  byte-equivalent intermediate reports.

CI performs a second process-level replay, keeps output under the runner
temporary directory, and uploads the first report under its exact head SHA.
