package generator

const (
	canonicalMarkerManifestSchemaV1       = "gooo.generator.marker-manifest"
	canonicalMarkerManifestProfileV1      = "generated-regions-and-slots"
	canonicalMarkerManifestVersionV1      = 1
	canonicalMarkerManifestEncodingV1     = "utf-8-byte-length-hex"
	canonicalMarkerManifestNewlineV1      = "LF"
	canonicalMarkerManifestRegionRecordV1 = "region"
	canonicalMarkerManifestSlotRecordV1   = "slot"
)

// markerManifestRegionV1 and markerManifestSlotV1 are the fixed typed fields
// emitted by canonicalMarkerManifestV1. IDs and kinds are source marker
// strings; their output tokens are UTF-8 byte length followed by lowercase
// hexadecimal bytes, so tabs, newlines, and map iteration cannot change the
// record grammar.
type markerManifestRegionV1 struct {
	ID        string
	Kind      string
	Start     int
	End       int
	StartLine int
	EndLine   int
	Slots     []markerManifestSlotV1
}
type markerManifestSlotV1 struct {
	ID         string
	RegionID   string
	RegionKind string
	Start      int
	End        int
	StartLine  int
	EndLine    int
}
