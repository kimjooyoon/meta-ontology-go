def known_verdict:
  . == "ACCEPT" or . == "REJECT";

def indicator($id; $class; $proof; $meta_operation; $observed; $expected; $audiences):
  {
    id: $id,
    class: $class,
    proof_choice: $proof,
    meta_operation: $meta_operation,
    observed: $observed,
    expected: $expected,
    satisfied: ($observed == $expected),
    audiences: $audiences
  };

def audience_view($indicators; $audience; $resolution):
  [$indicators[] | select(.audiences | index($audience))] as $visible
  | ($visible | length) as $total
  | ([$visible[] | select(.satisfied)] | length) as $satisfied
  | {
      audience: $audience,
      resolution: $resolution,
      satisfied: $satisfied,
      total: $total,
      basis_points: (if $total == 0 then 0 else (($satisfied * 10000 / $total) | floor) end)
    };

. as $input
| ($input.injected_observed | known_verdict) as $observed_known
| (if $observed_known and ($input.injected_observed == $input.source_observation.expected)
   then "FIXED_POINT"
   else "FAIL_CLOSED"
   end) as $decision
| (if $observed_known then "EXACT" else "LOWER_RESOLUTION" end) as $resolution
| (if ($observed_known | not)
   then "GENERATED_CONFORMANCE_DECISION_UNKNOWN"
   elif $decision == "FIXED_POINT"
   then "GENERATED_CONFORMANCE_DECISION_CONFIRMED"
   else "GENERATED_CONFORMANCE_DECISION_MISMATCH"
   end) as $reason
| [
    indicator(
      "source.external-decisions";
      "OUTCOME";
      "FOUNDATION";
      "count-source-external-decisions";
      $input.source_external_decisions;
      2;
      ["TOOL_AUTHOR", "GOVERNOR"]
    ),
    indicator(
      "counterfactual.unknown-injections";
      "DRIVER";
      "REGRESSION";
      "inject-unknown-into-generated-observation";
      1;
      1;
      ["TOOL_AUTHOR", "GOVERNOR"]
    ),
    indicator(
      "system.fail-closed-decisions";
      "OUTCOME";
      "FOUNDATION";
      "count-fail-closed-decisions";
      (if $decision == "FAIL_CLOSED" then 1 else 0 end);
      1;
      ["USER", "TOOL_AUTHOR", "GOVERNOR"]
    ),
    indicator(
      "system.lower-resolution-decisions";
      "OUTCOME";
      "COHERENCE";
      "count-lower-resolution-decisions";
      (if $resolution == "LOWER_RESOLUTION" then 1 else 0 end);
      1;
      ["USER", "TOOL_AUTHOR", "GOVERNOR"]
    ),
    indicator(
      "guardrail.false-fixed-points";
      "GUARDRAIL";
      "REGRESSION";
      "count-unknown-fixed-point-classifications";
      (if (($observed_known | not) and ($decision == "FIXED_POINT")) then 1 else 0 end);
      0;
      ["USER", "TOOL_AUTHOR", "GOVERNOR"]
    ),
    indicator(
      "guardrail.repository-writes";
      "GUARDRAIL";
      "FOUNDATION";
      "count-declared-repository-output-paths";
      $input.effects.repository_writes;
      0;
      ["GOVERNOR"]
    ),
    indicator(
      "guardrail.mutation-authorities";
      "GUARDRAIL";
      "COHERENCE";
      "join-counterfactual-mutation-authority";
      (if $input.effects.mutation_authority then 1 else 0 end);
      0;
      ["GOVERNOR"]
    )
  ] as $indicators
| ($indicators | length) as $total
| ([$indicators[] | select(.satisfied)] | length) as $satisfied
| (["FOUNDATION", "COHERENCE", "REGRESSION"] | map(
    . as $proof
    | [$indicators[] | select(.proof_choice == $proof)] as $selected
    | {
        proof_choice: $proof,
        satisfied: ([$selected[] | select(.satisfied)] | length),
        total: ($selected | length)
      }
  )) as $proofs
| {
    schema: "gooo/generated-unknown-resolution-report/v1",
    subject_sha: $input.subject_sha,
    metric_id: "gooo.metric.user.generated-unknown-resolution.v1",
    decision: $decision,
    resolution: $resolution,
    reason: $reason,
    source: {
      report_schema: $input.source_report_schema,
      report_digest: $input.source_report_digest,
      artifact_digest: $input.source_artifact_digest,
      validator_digest: $input.source_validator_digest,
      observation_id: $input.source_observation.id
    },
    counterfactual: {
      expected: $input.source_observation.expected,
      original_observed: $input.source_observation.observed,
      injected_observed: $input.injected_observed,
      observed_known: $observed_known
    },
    conformance: {
      decision: (if $satisfied == $total then "PASS" else "FAIL_CLOSED" end),
      coordinates: {
        satisfied: $satisfied,
        total: $total,
        basis_points: (if $total == 0 then 0 else (($satisfied * 10000 / $total) | floor) end)
      }
    },
    indicators: $indicators,
    views: [
      audience_view($indicators; "USER"; "USER_VISIBLE"),
      audience_view($indicators; "TOOL_AUTHOR"; "TOOL_CONTRACT"),
      audience_view($indicators; "GOVERNOR"; "FULL_RECEIPT")
    ],
    proofs: $proofs,
    promotion_credit_bps: 0,
    repository_writes: $input.effects.repository_writes,
    mutation_authority: $input.effects.mutation_authority,
    not_claimed: [
      "compiler-generated unknown verdict",
      "unknown recovery",
      "value-level execution",
      "domain correctness",
      "production readiness",
      "generalized verdict algebra"
    ]
  }
