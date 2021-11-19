module github.com/paketo-buildpacks/packit

go 1.16

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/anchore/syft v0.30.1
	github.com/cheggaaa/pb/v3 v3.0.8
	github.com/dsnet/compress v0.0.1
	github.com/gabriel-vasile/mimetype v1.4.0
	github.com/onsi/gomega v1.17.0
	github.com/pelletier/go-toml v1.9.4
	github.com/sclevine/spec v1.4.0
	github.com/ulikunitz/xz v0.5.10
)

// TODO: remove once https://github.com/anchore/syft/pull/635 is merged
replace github.com/anchore/syft => github.com/jonasagx/syft v0.27.1-0.20211118073839-eee29112ef6a
