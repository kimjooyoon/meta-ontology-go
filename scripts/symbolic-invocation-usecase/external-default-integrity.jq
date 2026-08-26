def expected_paths:
  [
    "external-default-input.json",
    "external-default-structural-validation.txt",
    "external-default-envelope.json",
    "external-default-observation.json",
    "external-default-coverage-report.json"
  ];

def digest_binding($input; $path):
  ([$input.files[] | select(.path == $path and (.digest | test("^sha256:[0-9a-f]{64}$")))] | length) as $observed
  | {
      id: ("guardrail.payload-digest." + $path),
      class: "GUARDRAIL",
      proof_choice: "FOUNDATION",
      meta_operation: "bind-external-default-coverage-payload",
      path: $path,
      observed: $observed,
      expected: 1,
      satisfied: ($observed == 1)
    };

. as $input
| expected_paths as $expected_paths
| (($input.files | length) == ($expected_paths | length)
    and ([$input.files[].path] | unique | length) == ($expected_paths | length)) as $shape_exact
| ([$expected_paths[] as $path | digest_binding($input; $path)] + [
    {
      id: "guardrail.manifest-entry-shape",
      class: "GUARDRAIL",
      proof_choice: "COHERENCE",
      meta_operation: "compare-exact-default-coverage-manifest",
      observed: (if $shape_exact then 1 else 0 end),
      expected: 1,
      satisfied: $shape_exact
    }
  ]) as $indicators
| ($indicators | length) as $total
| ([$indicators[] | select(.satisfied)] | length) as $satisfied
| {
    schema: "gooo/external-default-integrity-receipt/v1",
    subject_sha: $input.subject_sha,
    metric_id: "gooo.metric.guardrail.external-default-integrity.v1",
    decision: (if $satisfied == $total then "PASS" else "FAIL_CLOSED" end),
    resolution: (if $satisfied == $total then "EXACT" else "INVARIANT_ONLY" end),
    reason: (if $satisfied == $total
      then "EXTERNAL_DEFAULT_COVERAGE_PAYLOADS_BOUND"
      else "EXTERNAL_DEFAULT_COVERAGE_BINDING_INCOMPLETE"
      end),
    manifest: {path: $input.manifest_path, payload_bindings: ($input.files | length)},
    files: $input.files,
    coordinates: {
      satisfied: $satisfied,
      total: $total,
      basis_points: (if $total == 0 then 0 else (($satisfied * 10000 / $total) | floor) end)
    },
    indicators: $indicators,
    views: [
      {audience: "TOOL_AUTHOR", resolution: "TOOL_CONTRACT", satisfied: $satisfied, total: $total, basis_points: (if $total == 0 then 0 else (($satisfied * 10000 / $total) | floor) end)},
      {audience: "GOVERNOR", resolution: "FULL_RECEIPT", satisfied: $satisfied, total: $total, basis_points: (if $total == 0 then 0 else (($satisfied * 10000 / $total) | floor) end)}
    ],
    proofs: [
      {proof_choice: "FOUNDATION", satisfied: ([$indicators[] | select(.proof_choice == "FOUNDATION" and .satisfied)] | length), total: ([$indicators[] | select(.proof_choice == "FOUNDATION")] | length)},
      {proof_choice: "COHERENCE", satisfied: ([$indicators[] | select(.proof_choice == "COHERENCE" and .satisfied)] | length), total: ([$indicators[] | select(.proof_choice == "COHERENCE")] | length)},
      {proof_choice: "REGRESSION", satisfied: 0, total: 0}
    ],
    promotion_credit_bps: 0,
    repository_writes: $input.effects.repository_writes,
    mutation_authority: $input.effects.mutation_authority,
    not_claimed: [
      "central manifest inclusion",
      "integrity receipt self-binding",
      "production artifact signing"
    ]
  }
