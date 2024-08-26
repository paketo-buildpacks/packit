package model

import "github.com/anchore/syft/syft/formats/syftjson/model"

// Document represents the syft cataloging findings as a JSON document
type Document struct {
	Artifacts             []model.Package      `json:"artifacts"` // Artifacts is the list of packages discovered and placed into the catalog
	ArtifactRelationships []model.Relationship `json:"artifactRelationships"`
	Files                 []model.File         `json:"files,omitempty"`   // note: must have omitempty
	Secrets               []model.Secrets      `json:"secrets,omitempty"` // note: must have omitempty
	Source                Source               `json:"source"`            // Source represents the original object that was cataloged
	Distro                model.LinuxRelease   `json:"distro"`            // Distro represents the Linux distribution that was detected from the source
	Descriptor            model.Descriptor     `json:"descriptor"`        // Descriptor is a block containing self-describing information about syft
	Schema                model.Schema         `json:"schema"`            // Schema is a block reserved for defining the version for the shape of this JSON document and where to find the schema document to validate the shape
}

// Descriptor describes what created the document as well as surrounding metadata
type Descriptor struct {
	Name          string      `json:"name"`
	Version       string      `json:"version"`
	Configuration interface{} `json:"configuration,omitempty"`
}

type Schema struct {
	Version string `json:"version"`
	URL     string `json:"url"`
}
