package internal

// Copied from syft v0.42.3
// https://github.com/anchore/syft/blob/cc2c0e57a0d02a1719b4e34d0793f09e9699c8b0/internal/constants.go
const (
	// ApplicationName is the non-capitalized name of the application (do not change this)
	ApplicationName = "syft"

	// JSONSchemaVersion is the current schema version output by the JSON encoder
	// This is roughly following the "SchemaVer" guidelines for versioning the JSON schema. Please see schema/json/README.md for details on how to increment.
	JSONSchemaVersion = "3.1.1"
)
