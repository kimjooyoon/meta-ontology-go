# Compare public graph content with the existing profile witness, not case labels.
# This proves canonical-fixture structural parity only, not business execution.
def verdict($decision; $reason; $mismatches):
  {schema: "gooo/entity-fields-graph-comparison/v1", decision: $decision,
   scope: "CANONICAL_FIXTURE_STRUCTURAL_PARITY", reason: $reason,
   source_digest: $source_digest, mismatches: $mismatches,
   unknown: (if $decision == "UNKNOWN" then
     {stage: "GRAPH_SUPPORT", step: "LOAD_OBSERVATIONS", reason: $reason,
      unknown_class: "DIRECT_MISSING", next_operation: "RUN_CANONICAL_GRAPH_OBSERVATION", blocked_by: []}
     else null end)};
def position: {offset, line, column};
def witness_position: {offset: .Offset, line: .Line, column: .Column};
def graph_entities:
  [.nodes[] | select(.kind == "Entity") |
   {id, name, fields: [(.fields // [])[] |
     {id, parent, name, type_ref_id, presence, cardinality,
      source: {file: .source.file, start: (.source.start | position), end: (.source.end | position)}}]}]
  | sort_by(.id);
def witness_entities:
  [.Entities[] | {id: .ID, name: .Name, fields: [(.Fields // [])[] |
    {id: .ID, parent: .Parent, name: .Name, type_ref_id: .TypeRefID,
     presence: .Presence, cardinality: .Cardinality,
     source: {file: .Source.URI, start: (.Source.Start | witness_position), end: (.Source.End | witness_position)}}]}]
  | sort_by(.id);
def graph_activities:
  [.nodes[] | select(.kind == "Activity") | {id, name}] | sort_by(.id);
def witness_activities:
  [.Activities[] | {id: .ID, name: .Name}] | sort_by(.id);
def graph_relations:
  [.relations[] | {subject, predicate, object, status}]
  | sort_by([.subject, .predicate, .object, .status]);
def witness_relations:
  [.Activities[] | . as $activity |
   ((.Inputs // [])[] | {subject: $activity.ID, predicate: "used", object: .EntityID, status: "deterministic"}),
   ((.Outputs // [])[] | {subject: .EntityID, predicate: "wasGeneratedBy", object: $activity.ID, status: "deterministic"})]
  | sort_by([.subject, .predicate, .object, .status]);
def raw_digest: type == "string" and test("^[0-9a-f]{64}$");
if ($graph | length) == 0 or ($witness | length) == 0 then
  verdict("UNKNOWN"; "GRAPH_OR_WITNESS_OBSERVATION_MISSING"; [])
elif ($graph | length) != 1 or ($witness | length) != 1 then
  verdict("REFUTED"; "OBSERVATION_RECORD_COUNT_MISMATCH"; ["record_count"])
else
  $graph[0] as $g | $witness[0] as $w |
  if ($g | type) != "object" or ($w | type) != "object" then
    verdict("REFUTED"; "OBSERVATION_SHAPE_INVALID"; ["object_shape"])
  elif ($g.nodes | type) != "array" or ($g.relations | type) != "array"
       or ($w.Entities | type) != "array" or ($w.Activities | type) != "array" then
    verdict("REFUTED"; "OBSERVATION_SHAPE_INVALID"; ["collection_shape"])
  else
    [if $g.schema_version != "gooo-graph/v1" then "graph_schema" else empty end,
     if $g.source_digest != $source_digest then "source_digest" else empty end,
     if ($source_digest | raw_digest | not) then "input_digest_shape" else empty end,
     if $g.authorities.graph != "derived" then "graph_authority" else empty end,
     if $g.ir.status != "available" then "ir_status" else empty end,
     if $g.lowering.status != "available" then "lowering_status" else empty end,
     if ($g.graph_hash | raw_digest | not) then "graph_hash_shape" else empty end,
     if ($g.ir.semantic_digest | raw_digest | not) then "semantic_digest_shape" else empty end,
     if any($g.nodes[]; .kind != "Entity" and .kind != "Activity") then "unexpected_node_kind" else empty end,
     if ($g | graph_entities) != ($w | witness_entities) then "entity_field_projection" else empty end,
     if ($g | graph_activities) != ($w | witness_activities) then "activity_projection" else empty end,
     if ($g | graph_relations) != ($w | witness_relations) then "relation_projection" else empty end] as $mismatches |
    verdict((if ($mismatches | length) == 0 then "PASS" else "REFUTED" end);
      (if ($mismatches | length) == 0 then "CANONICAL_GRAPH_STRUCTURAL_PARITY" else "GRAPH_WITNESS_CONTRADICTION" end); $mismatches)
    + {compared: {entities: ($w.Entities | length), fields: ([$w.Entities[].Fields[]?] | length),
                 activities: ($w.Activities | length), relations: ($w | witness_relations | length)},
       excluded_claims: ["namespace parity", "independent semantic oracle", "business behavior", "semantic hash equivalence", "language completeness"]}
  end
end
