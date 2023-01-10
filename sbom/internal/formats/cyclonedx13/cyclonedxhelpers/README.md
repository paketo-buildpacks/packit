# Source
The contents of this directory is largely based on anchore/syft's
internal `cyclonedxhelpers` package. The version copied is from an [old
commit](https://github.com/anchore/syft/blob/a86dd3704efdb19aea22774eb7e099d4e85d41e4/internal/formats/common/cyclonedxhelpers)
of Syft that supports CycloneDX JSON Schema 1.3.

Any helpers here remain because they contain 1.3-specific logic, so we cannot
use upstream code.

The implementation of `decoder` has been omitted for
simplicity, since it is not required for buildpacks' SBOM generation.

