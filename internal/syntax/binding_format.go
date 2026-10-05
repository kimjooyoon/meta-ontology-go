package syntax

import (
	"fmt"
	"strings"
)

func formatBinding(output *strings.Builder, binding BindingDecl) error {
	if binding.Producer.PackageAlias != "" {
		if err := validateIdentifier(binding.Producer.PackageAlias, "binding producer package alias"); err != nil {
			return err
		}
	}
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
	keyword := "bind"
	if binding.Feedback {
		keyword = "feedback"
	}
	fmt.Fprintf(output, "%s %s%s.%s -> %s%s.%s", keyword,
		qualifiedPrefix(binding.Producer.PackageAlias), binding.Producer.Activity.Name, binding.Producer.Port.Name,
		qualifiedPrefix(binding.Consumer.PackageAlias), binding.Consumer.Activity.Name, binding.Consumer.Port.Name)
	return nil
}

func qualifiedPrefix(alias string) string {
	if alias == "" {
		return ""
	}
	return alias + "."
}
