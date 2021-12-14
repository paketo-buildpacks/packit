module github.com/paketo-buildpacks/packit/v2

go 1.16

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/buildpacks/libcnb v1.24.1-0.20211118224334-5beb81e48b20
	github.com/dsnet/compress v0.0.2-0.20210315054119-f66993602bf5
	github.com/gabriel-vasile/mimetype v1.4.0
	github.com/kr/text v0.2.0 // indirect
	github.com/onsi/gomega v1.17.0
	github.com/pelletier/go-toml v1.9.4
	github.com/sclevine/spec v1.4.0
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/ulikunitz/xz v0.5.10
	golang.org/x/net v0.0.0-20211111160137-58aab5ef257a // indirect
	golang.org/x/sys v0.0.0-20211110154304-99a53858aa08 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

// TODO: remove once a new release is cut with the CycloneDX format (probably v0.32.0)
replace github.com/anchore/syft => github.com/anchore/syft v0.31.1-0.20211204010623-5374a1dc6ff6
