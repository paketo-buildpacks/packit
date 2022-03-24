# Source
The contents of this directory is largely based on anchore/syft's
internal `cyclonedx13json` package. The version copied is from an [old
commit](https://github.com/anchore/syft/blob/a86dd3704efdb19aea22774eb7e099d4e85d41e4/internal/formats/cyclonedx13json)
of Syft that supports CycloneDX JSON Schema 1.3.

The implementations of `decoder` and `validator` have been omitted for
simplicity, since they are not required for buildpacks' SBOM generation.

