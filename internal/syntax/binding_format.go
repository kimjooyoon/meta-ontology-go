package syntax

import (
	"fmt"
	"strings"
)

func formatBinding(output *strings.Builder, binding BindingDecl) error {
	if err := validateIdentifier(binding.Producer.Activity.Name, "binding producer activity"); err != nil {
		return err
	}
	if err := validateIdentifier(binding.Producer.Port.Name, "binding producer port"); err != nil {
		return err
	}
	if err := validateIdentifier(binding.Consumer.Activity.Name, "binding consumer activity"); err != nil {
		return err
	}
	if err := validateIdentifier(binding.Consumer.Port.Name, "binding consumer port"); err != nil {
		return err
	}
	fmt.Fprintf(output, "bind %s.%s -> %s.%s", binding.Producer.Activity.Name, binding.Producer.Port.Name, binding.Consumer.Activity.Name, binding.Consumer.Port.Name)
	return nil
}
