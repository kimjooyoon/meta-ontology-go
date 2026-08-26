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
| ($input.envelopes | length) as $value_projections
| ([$input.envelopes[] | select(.decision == "READY")] | length) as $ready_envelopes
| ([$input.envelopes[] | select(.decision == "FAIL_CLOSED" and .resolution == "LOWER_RESOLUTION")] | length) as $failed_envelopes
| ([$input.observations[]
    | select(.decision == "OBSERVED_READY" or .decision == "OBSERVED_FAIL_CLOSED")] | length) as $consumer_observations
| ([$input.replay_matches[] | select(. == true)] | length) as $replay_matches
| ([$input.envelopes[].effects.requested_invocations] | add // 0) as $requested_invocations
| ([$input.envelopes[].effects.executed_invocations, $input.observations[].effects.executed_invocations] | add // 0) as $executed_invocations
| ([$input.envelopes[].effects.repository_writes, $input.observations[].effects.repository_writes] | add // 0) as $repository_writes
| ([$input.envelopes[].effects.mutation_authority, $input.observations[].effects.mutation_authority]
    | map(select(. == true)) | length) as $mutation_authorities
| [
    indicator(
      "source.gooo-files";
      "DRIVER";
      "FOUNDATION";
      "count-compiler-source-gooo-files";
      $input.source.gooo_files;
      3;
      ["TOOL_AUTHOR", "GOVERNOR"]
    ),
    indicator(
      "source.go-files";
      "GUARDRAIL";
      "FOUNDATION";
      "count-compiler-source-go-files";
      $input.source.go_files;
      0;
      ["TOOL_AUTHOR", "GOVERNOR"]
    ),
    indicator(
      "source.gooo-lines";
      "DRIVER";
      "FOUNDATION";
      "sum-compiler-source-gooo-lines";
      $input.source.gooo_lines;
      16;
      ["TOOL_AUTHOR", "GOVERNOR"]
    ),
    indicator(
      "source.generated-vectors";
      "DRIVER";
      "FOUNDATION";
      "count-compiler-generated-value-vectors";
      $input.source.generated_vectors;
      2;
      ["TOOL_AUTHOR", "GOVERNOR"]
    ),
    indicator(
      "source.external-decisions";
      "DRIVER";
      "COHERENCE";
      "count-external-vector-decisions";
      $input.source.external_decisions;
      2;
      ["TOOL_AUTHOR", "GOVERNOR"]
    ),
    indicator(
      "source.compiler-value-contract-bindings";
      "DRIVER";
      "FOUNDATION";
      "bind-compiler-symbolic-value-contract";
      $input.source.contract_bindings;
      1;
      ["TOOL_AUTHOR", "GOVERNOR"]
    ),
    indicator(
      "user.value-projections";
      "OUTCOME";
      "COHERENCE";
      "project-generated-vectors-to-value-envelopes";
      $value_projections;
      2;
      ["USER", "TOOL_AUTHOR", "GOVERNOR"]
    ),
    indicator(
      "user.ready-envelopes";
      "OUTCOME";
      "FOUNDATION";
      "count-ready-value-envelopes";
      $ready_envelopes;
      1;
      ["USER", "TOOL_AUTHOR", "GOVERNOR"]
    ),
    indicator(
      "user.fail-closed-envelopes";
      "OUTCOME";
      "REGRESSION";
      "count-lower-resolution-value-envelopes";
      $failed_envelopes;
      1;
      ["USER", "TOOL_AUTHOR", "GOVERNOR"]
    ),
    indicator(
      "user.consumer-observations";
      "OUTCOME";
      "COHERENCE";
      "reduce-value-envelopes-for-user";
      $consumer_observations;
      2;
      ["USER", "TOOL_AUTHOR", "GOVERNOR"]
    ),
    indicator(
      "tool.deterministic-replay-matches";
      "DRIVER";
      "REGRESSION";
      "compare-projection-and-consumer-replays";
      $replay_matches;
      4;
      ["TOOL_AUTHOR", "GOVERNOR"]
    ),
    indicator(
      "effect.requested-invocations";
      "DRIVER";
      "COHERENCE";
      "sum-requested-symbolic-invocations";
      $requested_invocations;
      1;
      ["USER", "TOOL_AUTHOR", "GOVERNOR"]
    ),
    indicator(
      "guardrail.executed-invocations";
      "GUARDRAIL";
      "REGRESSION";
      "sum-executed-symbolic-invocations";
      $executed_invocations;
      0;
      ["USER", "TOOL_AUTHOR", "GOVERNOR"]
    ),
    indicator(
      "guardrail.repository-writes";
      "GUARDRAIL";
      "FOUNDATION";
      "sum-value-projection-repository-writes";
      $repository_writes;
      0;
      ["GOVERNOR"]
    ),
    indicator(
      "guardrail.mutation-authorities";
      "GUARDRAIL";
      "COHERENCE";
      "join-value-projection-mutation-authority";
      $mutation_authorities;
      0;
      ["GOVERNOR"]
    )
  ] as $indicators
| ($indicators | length) as $total
| ([$indicators[] | select(.satisfied)] | length) as $satisfied
| (["OUTCOME", "DRIVER", "GUARDRAIL"] | map(
    . as $class
    | [$indicators[] | select(.class == $class)] as $selected
    | {
        class: $class,
        satisfied: ([$selected[] | select(.satisfied)] | length),
        total: ($selected | length)
      }
  )) as $classes
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
    schema: "gooo/generated-value-projection-report/v1",
    subject_sha: $input.subject_sha,
    metric_id: "gooo.metric.user.generated-value-projection.v1",
    decision: (if $satisfied == $total then "PASS" else "FAIL_CLOSED" end),
    resolution: (if $satisfied == $total then "VALUE_PROJECTION_ONLY" else "INVARIANT_ONLY" end),
    reason: (if $satisfied == $total
      then "GENERATED_VALUES_PROJECTED_AND_OBSERVED"
      else "GENERATED_VALUE_PROJECTION_INCOMPLETE"
      end),
    source: $input.source,
    coordinates: {
      satisfied: $satisfied,
      total: $total,
      basis_points: (if $total == 0 then 0 else (($satisfied * 10000 / $total) | floor) end)
    },
    classes: $classes,
    indicators: $indicators,
    views: [
      audience_view($indicators; "USER"; "USER_VISIBLE"),
      audience_view($indicators; "TOOL_AUTHOR"; "TOOL_CONTRACT"),
      audience_view($indicators; "GOVERNOR"; "FULL_RECEIPT")
    ],
    proofs: $proofs,
    envelopes: [$input.envelopes[] | {
      vector_id: .source.vector_id,
      decision,
      resolution,
      reason,
      invocation,
      diagnostics,
      digest
    }],
    observations: [$input.observations[] | {
      vector_id: .source.vector_id,
      decision,
      resolution,
      reason,
      value,
      digest
    }],
    effects: {
      requested_invocations: $requested_invocations,
      executed_invocations: $executed_invocations,
      repository_writes: $repository_writes,
      mutation_authorities: $mutation_authorities
    },
    promotion_credit_bps: 0,
    repository_writes: $repository_writes,
    mutation_authority: ($mutation_authorities > 0),
    not_claimed: [
      "effect execution",
      "arbitrary user input",
      "complete interpreter semantics",
      "domain correctness",
      "production readiness",
      "performance beyond these fixed vectors"
    ]
  }
