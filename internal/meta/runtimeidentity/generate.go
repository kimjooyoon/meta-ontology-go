package runtimeidentity

//go:generate go run ../../../scripts/library-projection -root ../../.. -mode write -source internal/meta/runtimeidentity/fields.gooo -out internal/meta/runtimeidentity/semantic.gooo.go -entry urn:gooo:runtime-identity:describe -go-package runtimeidentity
