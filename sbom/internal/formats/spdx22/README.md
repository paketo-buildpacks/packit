# Source
The contents of this directory is largely based on anchore/syft's `spdxjson`
package. The version copied is from the (current) latest release of Syft,
[v0.65.0](https://github.com/anchore/syft/tree/v0.65.0/syft/formats/spdxjson),
which currently supports SPDX 2.3. This choice was made so that we can leverage
as much as the common Syft helper functionality (such as
[spdxhelpers](https://github.com/anchore/syft/tree/v0.65.0/syft/formats/common/spdxhelpers))
as possible.

Modifications to the code have been made to the code, and to the code in the
`model` directory to include subtle differences that apply to SPDX 2.2. These
changes were largely based on
[v0.60.3](https://github.com/anchore/syft/tree/v0.60.3/syft/formats/spdx22).

The implementations of `decoder` and `validator` have been omitted for
simplicity, since they are not required for buildpacks' SBOM generation.
