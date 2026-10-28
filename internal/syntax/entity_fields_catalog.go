package syntax

// Catalog order is the kernel's supported profile order. The canonical lookup
// and advisory consumers share the same immutable identities and digests.
var supportedEntityFieldsProfiles = [...]EntityFieldsProfile{
	EntityFieldsV1Support().Profile,
	EntityFieldsV2Support().Profile,
	EntityFieldsV3Support().Profile,
	EntityFieldsV4Support().Profile,
	EntityFieldsV5Support().Profile,
	EntityFieldsV6Support().Profile,
	EntityFieldsV7Support().Profile,
}

// SupportedEntityFieldsProfiles returns detached values, not the kernel table.
func SupportedEntityFieldsProfiles() []EntityFieldsProfile {
	return append([]EntityFieldsProfile(nil), supportedEntityFieldsProfiles[:]...)
}
