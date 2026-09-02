def digest_binding($input; $id; $path; $meta_operation; $proof):
  ([$input.files[]
    | select(
        .path == $path
        and (.digest | test("^sha256:[0-9a-f]{64}$"))
      )] | length) as $observed
  | {
      id: $id,
      class: "GUARDRAIL",
      proof_choice: $proof,
      meta_operation: $meta_operation,
      observed: $observed,
      expected: 1,
      satisfied: ($observed == 1)
    };

. as $input
| (($input.files | length) == 2 and ([$input.files[].path] | unique | length) == 2) as $shape_exact
| [
    digest_binding(
      $input;
      "guardrail.counterfactual-digest-bindings";
      "generated-unknown-counterfactual.json";
      "bind-counterfactual-file-digest";
      "FOUNDATION"
    ),
    digest_binding(
      $input;
      "guardrail.report-digest-bindings";
      "generated-unknown-resolution-report.json";
      "bind-resolution-report-file-digest";
      "FOUNDATION"
    ),
    {
      id: "guardrail.manifest-entry-shape",
      class: "GUARDRAIL",
      proof_choice: "COHERENCE",
      meta_operation: "compare-exact-manifest-entry-set",
      observed: (if $shape_exact then 1 else 0 end),
      expected: 1,
      satisfied: $shape_exact
    }
  ] as $indicators
| ($indicators | length) as $total
| ([$indicators[] | select(.satisfied)] | length) as $satisfied
| {
    schema: "gooo/generated-unknown-integrity-receipt/v1",
    subject_sha: $input.subject_sha,
    metric_id: "gooo.metric.guardrail.generated-unknown-integrity.v1",
    decision: (if $satisfied == $total then "PASS" else "FAIL_CLOSED" end),
    resolution: (if $satisfied == $total then "EXACT" else "INVARIANT_ONLY" end),
    reason: (if $satisfied == $total
      then "GENERATED_UNKNOWN_PAYLOADS_BOUND"
      else "GENERATED_UNKNOWN_PAYLOAD_BINDING_INCOMPLETE"
      end),
    manifest: {
      path: $input.manifest_path,
      payload_bindings: ($input.files | length)
    },
    files: $input.files,
    coordinates: {
      satisfied: $satisfied,
      total: $total,
      basis_points: (if $total == 0 then 0 else (($satisfied * 10000 / $total) | floor) end)
    },
    indicators: $indicators,
    views: [
      {
        audience: "TOOL_AUTHOR",
        resolution: "TOOL_CONTRACT",
        satisfied: $satisfied,
        total: $total,
        basis_points: (if $total == 0 then 0 else (($satisfied * 10000 / $total) | floor) end)
      },
      {
        audience: "GOVERNOR",
        resolution: "FULL_RECEIPT",
        satisfied: $satisfied,
        total: $total,
        basis_points: (if $total == 0 then 0 else (($satisfied * 10000 / $total) | floor) end)
      }
    ],
    proofs: [
      {
        proof_choice: "FOUNDATION",
        satisfied: ([$indicators[] | select(.proof_choice == "FOUNDATION" and .satisfied)] | length),
        total: ([$indicators[] | select(.proof_choice == "FOUNDATION")] | length)
      },
      {
        proof_choice: "COHERENCE",
        satisfied: ([$indicators[] | select(.proof_choice == "COHERENCE" and .satisfied)] | length),
        total: ([$indicators[] | select(.proof_choice == "COHERENCE")] | length)
      },
      {
        proof_choice: "REGRESSION",
        satisfied: 0,
        total: 0
      }
    ],
    promotion_credit_bps: 0,
    repository_writes: $input.effects.repository_writes,
    mutation_authority: $input.effects.mutation_authority,
    not_claimed: [
      "central manifest inclusion",
      "integrity receipt self-binding",
      "provenance beyond the exact head",
      "production artifact signing"
    ]
  }
