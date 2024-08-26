# Source
The contents of this directory is largely based on anchore/syft's internal
`syftjson` package. The version copied is from
[v0.41.1](https://github.com/anchore/syft/blob/07d3c9af52f241613971ccadd18c8f8ab67abc4e)
of Syft that supports Syft JSON Schema 3.0.1.

The implementations of `decoder` and `validator` have been omitted for
simplicity, since they are not required for buildpacks' SBOM generation.

Aspects of the model have been copied over due to slight deviations against the
latest Syft JSON model.
