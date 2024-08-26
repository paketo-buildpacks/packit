package sbom_test

import "time"

/* A set of structs that are used to unmarshal SBOM JSON output in tests */

type license struct {
	License struct {
		ID string `json:"id"`
	} `json:"license"`
}

type component struct {
	Type     string    `json:"type"`
	Name     string    `json:"name"`
	Version  string    `json:"version"`
	Licenses []license `json:"licenses"`
	PURL     string    `json:"purl"`
}

type cdxOutput struct {
	BOMFormat    string `json:"bomFormat"`
	SpecVersion  string `json:"specVersion"`
	SerialNumber string `json:"serialNumber"`
	Metadata     struct {
		Timestamp string `json:"timestamp"`
		Component struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"component"`
	} `json:"metadata"`
	Components []component `json:"components"`
}

type artifact struct {
	Name     string   `json:"name"`
	Version  string   `json:"version"`
	Licenses []string `json:"licenses"`
	CPEs     []string `json:"cpes"`
	PURL     string   `json:"purl"`
}

type syftOutput struct {
	Artifacts []artifact `json:"artifacts"`
	Source    struct {
		Type   string `json:"type"`
		Target string `json:"target"`
	} `json:"source"`
	Schema struct {
		Version string `json:"version"`
	} `json:"schema"`
}

type externalRef struct {
	Category string `json:"referenceCategory"`
	Locator  string `json:"referenceLocator"`
	Type     string `json:"referenceType"`
}

type pkg struct {
	ExternalRefs     []externalRef `json:"externalRefs"`
	LicenseConcluded string        `json:"licenseConcluded"`
	LicenseDeclared  string        `json:"licenseDeclared"`
	Name             string        `json:"name"`
	Version          string        `json:"versionInfo"`
}

type spdxOutput struct {
	Packages          []pkg  `json:"packages"`
	SPDXVersion       string `json:"spdxVersion"`
	DocumentNamespace string `json:"documentNamespace"`
	CreationInfo      struct {
		Created time.Time `json:"created"`
	} `json:"creationInfo"`
}
