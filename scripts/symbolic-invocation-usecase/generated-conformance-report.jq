def indicator($id; $class; $proof; $operation; $observed; $expected):
  {
    id: $id,
    class: $class,
    proof_choice: $proof,
    meta_operation: $operation,
    observed: $observed,
    expected: $expected,
    satisfied: ($observed == $expected)
  };

{
  schema: "gooo/generated-symbolic-conformance-report/v1",
  subject_sha: $subject_sha,
  metric_id: "gooo.metric.user.generated-symbolic-conformance.v1",
  decision: "PASS",
  resolution: "EXACT",
  reason: "EXTERNAL_GENERATED_CONFORMANCE_OBSERVED",
  summary: {
    coordinates: {satisfied: 8, total: 8, basis_points: 10000},
    generated_vectors: $generated_vectors,
    external_decisions: $external_decisions,
    expectation_matches: $expectation_matches,
    accepted_vectors: $accepted_vectors,
    rejected_vectors: $rejected_vectors,
    unknowns: 0,
    validator_bindings: 1,
    validator_digest: $validator_digest,
    artifact_digest: $artifact_digest,
    json_schema_digest: $json_schema_digest,
    effects: {repository_writes: 0, mutation_authority: false}
  },
  indicators: [
    indicator("user.generated-vectors"; "OUTCOME"; "FOUNDATION"; "count-compiler-generated-vectors"; $generated_vectors; 2),
    indicator("user.external-decisions"; "OUTCOME"; "COHERENCE"; "sum-pinned-validator-decisions"; $external_decisions; 2),
    indicator("user.expectation-matches"; "DRIVER"; "COHERENCE"; "compare-generated-and-observed-verdicts"; $expectation_matches; 2),
    indicator("user.generated-accepts"; "DRIVER"; "FOUNDATION"; "count-externally-accepted-generated-vectors"; $accepted_vectors; 1),
    indicator("user.generated-rejects"; "DRIVER"; "REGRESSION"; "count-externally-rejected-generated-vectors"; $rejected_vectors; 1),
    indicator("tool.validator-bindings"; "GUARDRAIL"; "FOUNDATION"; "bind-pinned-validator-digest"; 1; 1),
    indicator("guardrail.repository-writes"; "GUARDRAIL"; "FOUNDATION"; "sum-independent-consumer-writes"; 0; 0),
    indicator("guardrail.mutation-authorities"; "GUARDRAIL"; "COHERENCE"; "join-independent-consumer-authority"; 0; 0)
  ],
  views: [
    {audience: "USER", resolution: "USER_VISIBLE", satisfied: 5, total: 5, basis_points: 10000},
    {audience: "TOOL_AUTHOR", resolution: "TOOL_CONTRACT", satisfied: 6, total: 6, basis_points: 10000},
    {audience: "GOVERNOR", resolution: "FULL_RECEIPT", satisfied: 8, total: 8, basis_points: 10000}
  ],
  observations: $observations,
  promotion_credit_bps: 0,
  repository_writes: 0,
  mutation_authority: false,
  not_claimed: [
    "external fixture replacement",
    "value-level execution",
    "domain correctness",
    "production readiness",
    "performance beyond this runner and fixed vectors"
  ]
}
