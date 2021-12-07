module github.com/paketo-buildpacks/packit/v2

go 1.16

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/anchore/syft v0.31.0
	github.com/cheggaaa/pb/v3 v3.0.8
	github.com/dsnet/compress v0.0.2-0.20210315054119-f66993602bf5
	github.com/gabriel-vasile/mimetype v1.4.0
	github.com/onsi/gomega v1.17.0
	github.com/pelletier/go-toml v1.9.4
	github.com/sclevine/spec v1.4.0
	github.com/ulikunitz/xz v0.5.10
)

// TODO: remove once a new release is cut with the CycloneDX format (probably v0.32.0)
replace github.com/anchore/syft => github.com/anchore/syft v0.31.1-0.20211204010623-5374a1dc6ff6
