def activity_valid:
  if (.activity | type) != "string" then false else (.activity | length) > 0 end;

def inputs_valid:
  if (.inputs | type) != "array" then
    false
  else
    (.inputs | length) > 0
    and all(.inputs[]; ((type == "string") and (length > 0)))
  end;

def activity_matches($expected; $instance):
  if $expected == "NON_EMPTY" then
    ($instance | activity_valid)
  elif $expected == "MISSING_OR_EMPTY" then
    (($instance | activity_valid) | not)
  elif $expected == "ANY" then
    true
  else
    false
  end;

def inputs_matches($expected; $instance):
  if $expected == "NON_EMPTY" then
    ($instance | inputs_valid)
  elif $expected == "ANY" then
    true
  else
    false
  end;

def rule_matches($rule; $instance):
  activity_matches($rule.match.activity; $instance)
  and inputs_matches($rule.match.inputs; $instance);

. as $input
| $input.vector.instance as $instance
| (($instance | activity_valid) and ($instance | inputs_valid)) as $valid
| ([$input.contract.rules[] | select(rule_matches(.; $instance))] | first) as $matched_rule
| ($matched_rule // $input.contract.default) as $decision_rule
| {
    schema: "gooo/symbolic-invocation-value-envelope/v1",
    subject_sha: $input.subject_sha,
    language: "gooo",
    source: {
      artifact_digest: $input.source.artifact_digest,
      external_report_digest: $input.source.external_report_digest,
      validator_digest: $input.source.validator_digest,
      contract_digest: $input.contract.digest,
      rule_id: ($matched_rule.id // "default"),
      vector_id: $input.vector.id,
      expected: $input.vector.expected,
      proof_choice: $input.vector.proof_choice,
      meta_operation: $input.vector.meta_operation
    },
    decision: $decision_rule.decision,
    resolution: $decision_rule.resolution,
    reason: $decision_rule.reason,
    invocation: (if $decision_rule.decision == "READY" and $valid
      then {
        activity: $instance.activity,
        inputs: $instance.inputs,
        input_count: ($instance.inputs | length)
      }
      else {
        input_count: (if ($instance.inputs | type) == "array" then ($instance.inputs | length) else 0 end)
      }
      end),
    diagnostics: [
      if ($instance | activity_valid) then empty else "activity-required" end,
      if ($instance | inputs_valid) then empty else "inputs-required" end
    ],
    effects: {
      requested_invocations: (if $decision_rule.decision == "READY" and $valid then 1 else 0 end),
      executed_invocations: 0,
      repository_writes: $input.effects.repository_writes,
      mutation_authority: $input.effects.mutation_authority
    }
  }
