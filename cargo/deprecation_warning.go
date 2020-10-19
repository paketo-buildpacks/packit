package cargo

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type DeprecationWarning struct{}

func NewDeprecationWarning() DeprecationWarning {
	return DeprecationWarning{}
}

func (d DeprecationWarning) WarnDeprecatedFields(path string) error {
	var deprecatedFields struct {
		Metadata struct {
			IncludeFiles []string `toml:"include_files"`
			PrePackage   string   `toml:"pre_package"`
		} `toml:"metadata"`
	}

	_, err := toml.DecodeFile(path, &deprecatedFields)
	if err != nil {
		return err
	}

	if !(deprecatedFields.Metadata.IncludeFiles == nil && deprecatedFields.Metadata.PrePackage == "") {
		return fmt.Errorf("the include_files and pre_package fields in the metadata section of the buildpack.toml have been changed to include-files and pre-package respectively: please update the buildpack.toml to reflect this change")
	}

	return nil
}
