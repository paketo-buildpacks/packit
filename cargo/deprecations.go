package cargo

import (
	"fmt"
	"io"
	"sort"

	"github.com/BurntSushi/toml"
)

// Deprecation describes a deprecated dependency metadata key found in a
// buildpack.toml or extension.toml document.
type Deprecation struct {
	// Path is the location of the dependency in the document, for example
	// "metadata.dependencies[2]".
	Path string

	// ID is the identifier of the dependency, if one is set.
	ID string

	// Key is the deprecated key.
	Key string

	// Replacement describes what should be used instead of the deprecated key.
	Replacement string
}

// dependencyKeyDeprecations maps deprecated dependency metadata keys to what
// should be used instead.
//
// Only renames that are safe to apply are listed. sha256, stacks, and the
// legacy list-of-identifiers form of licenses are deliberately absent:
//
//   - sha256 and stacks are still written by tooling (including jam) as part of
//     dependency updates, and neither is marked deprecated on the config
//     structs.
//   - rewriting licenses from a list of identifiers to a list of tables breaks
//     parsing for any consumer still on a packit release that predates table
//     support, so it should not be encouraged until that support is widespread.
//
// Reporting them would produce warnings that cannot be acted on safely.
var dependencyKeyDeprecations = map[string]string{
	"cpe":              "cpes",
	"deprecation_date": "eol-date",
	"purl":             "purls",
	"source-sha256":    "source-checksum",
	"source_sha256":    "source-checksum",
}

// DeprecatedDependencyKeys returns the deprecated dependency metadata keys used
// in the given buildpack.toml or extension.toml document. This lets tooling warn
// authors that a document still uses a format which predates the standard
// dependency metadata format.
func DeprecatedDependencyKeys(r io.Reader) ([]Deprecation, error) {
	var document map[string]interface{}

	if _, err := toml.NewDecoder(r).Decode(&document); err != nil {
		return nil, err
	}

	metadata, ok := document["metadata"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	var deprecations []Deprecation

	for index, dependency := range dependencyTables(metadata["dependencies"]) {
		id, _ := dependency["id"].(string)

		keys := make([]string, 0, len(dependency))
		for key := range dependency {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			replacement, deprecated := dependencyKeyDeprecations[key]
			if !deprecated {
				continue
			}

			deprecations = append(deprecations, Deprecation{
				Path:        fmt.Sprintf("metadata.dependencies[%d]", index),
				ID:          id,
				Key:         key,
				Replacement: replacement,
			})
		}
	}

	return deprecations, nil
}

// dependencyTables normalizes the decoded dependencies into a list of tables.
func dependencyTables(value interface{}) []map[string]interface{} {
	switch dependencies := value.(type) {
	case []map[string]interface{}:
		return dependencies

	case []interface{}:
		tables := make([]map[string]interface{}, 0, len(dependencies))
		for _, dependency := range dependencies {
			if table, ok := dependency.(map[string]interface{}); ok {
				tables = append(tables, table)
			}
		}
		return tables

	default:
		return nil
	}
}
