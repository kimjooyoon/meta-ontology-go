# EntityFields V1 support view

This section is generated from the profile-bound observation and observed public routes; it is not a language-completeness score.

profile_id=gooo.entityfields.go-projection.v1
profile_version=1
meta_source=internal/meta/entityfields/entity-fields-meta.gooo
fixture=examples/entity-fields-v1/main.gooo

| Surface | State | Scope |
| --- | --- | --- |
| ordinary/default parser | DEFERRED | default parser remains deferred until explicit V1 opt-in |
| explicit V1 parser | SUPPORTED | profile-bound syntax parser on the canonical fixture |
| formatter | SUPPORTED | syntax formatting and canonical replay |
| semantic lowerer | SUPPORTED | profile-bound semantic IR lowering |
| BX Get/Put | SUPPORTED | source-preserving and semantic mutation round trips |
| Go generator | SUPPORTED | profile-bound structural Go projection |
| source map | SUPPORTED | generated structural source-map projection |
| LSP adapter | SUPPORTED | EntityFieldsSyntaxParser adapter evidence only |
| public CLI check | SUPPORTED | gooo check --semantic on the canonical fixture |
| public CLI generate | SUPPORTED | gooo generate and exact generated artifact comparison |
| generated Go | SUPPORTED | generated module compilation only, not business-behavior proof |
| LSP server/editor support | UNKNOWN | no server or editor integration claim without separate evidence |
