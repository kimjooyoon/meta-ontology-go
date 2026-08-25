package artifactemit

const PackageReceiptSchema = "gooo/package-source-execution-receipt/v1"

type packageReceipt struct {
	Schema      string             `json:"schema"`
	Decision    string             `json:"decision"`
	Resolution  string             `json:"resolution"`
	PackagePath string             `json:"package_path"`
	Package     string             `json:"package"`
	Namespace   string             `json:"namespace"`
	Entry       string             `json:"entry"`
	Sources     []sourceDefinition `json:"sources"`
	Execution   executionReceipt   `json:"execution"`
	Effects     Effects            `json:"effects"`
	Digest      string             `json:"digest"`
}

type executionReceipt struct {
	Entry operationEntry `json:"entry"`
}

type operationEntry struct {
	Package   string    `json:"package"`
	Namespace string    `json:"namespace"`
	Activity  string    `json:"activity"`
	Inputs    []Binding `json:"inputs"`
	Output    Binding   `json:"output"`
}

type sourceDefinition struct {
	Filename         string `json:"filename"`
	Digest           string `json:"digest"`
	DeclarationCount int    `json:"declaration_count"`
}
