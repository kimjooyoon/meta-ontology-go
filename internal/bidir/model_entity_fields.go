package bidir

import "github.com/kimjooyoon/meta-ontology-go/internal/semantic"

// ValidateWithEntityFieldsSupport validates the source-bound nominal registry
// as well as the model graph and field provenance, without changing either.
func (model Model) ValidateWithEntityFieldsSupport(support EntityFieldsSupport) error {
	if err := validateEntityFieldsSupport(support); err != nil {
		return err
	}
	registry, err := modelRecordRegistry(model.Namespace, model.Nodes, semantic.DefaultTypeRegistry(), support)
	if err != nil {
		return err
	}
	if err := model.ValidateWithTypes(registry); err != nil {
		return err
	}
	return validateEntityFieldsModel(model.Nodes, registry, support)
}
