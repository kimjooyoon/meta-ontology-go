package packageruntime

// Interface describes declarations in every package of a validated workspace.
// Activity bodies are never executed by Describe.
type Interface struct {
	Schema      string             `json:"schema"`
	Scope       string             `json:"scope"`
	Entry       EntrySpec          `json:"entry"`
	Packages    []InterfacePackage `json:"packages"`
	ImageDigest string             `json:"image_digest"`
	Digest      string             `json:"digest"`
}

type InterfacePackage struct {
	Path         string                 `json:"path"`
	Name         string                 `json:"name"`
	Namespace    string                 `json:"namespace"`
	Imports      []string               `json:"imports"`
	Sources      []SourceImage          `json:"sources"`
	Declarations []InterfaceDeclaration `json:"declarations"`
}

type InterfaceDeclaration struct {
	Kind   string           `json:"kind"`
	Name   string           `json:"name"`
	ID     string           `json:"id"`
	Source string           `json:"source"`
	Shape  string           `json:"shape,omitempty"`
	Fields []InterfaceField `json:"fields,omitempty"`
	Inputs []string         `json:"inputs,omitempty"`
	Output string           `json:"output,omitempty"`
}

type InterfaceField struct {
	Name        string   `json:"name"`
	ID          string   `json:"id"`
	Aliases     []string `json:"aliases,omitempty"`
	TypeID      string   `json:"type_id"`
	Presence    string   `json:"presence"`
	Cardinality string   `json:"cardinality"`
}

func Describe(manifest Manifest) (Interface, error) {
	image, packages, err := build(manifest, true)
	if err != nil {
		return Interface{}, err
	}
	result := Interface{
		Schema: "gooo/package-interface/v1", Scope: "DECLARED_PACKAGE_INTERFACES",
		Entry:    EntrySpec{PackagePath: image.Entry.PackagePath, Activity: image.Entry.Activity},
		Packages: packages, ImageDigest: image.Digest,
	}
	result.Digest = digestValue(result)
	return result, nil
}
