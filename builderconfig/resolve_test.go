package builderconfig_test

import (
	"fmt"
	"os"
	"path"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/packit/v2/builderconfig"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

type MockEnv map[string]string

func TestUnitConfig(t *testing.T) {
	suite := spec.New("packit/builderconfig", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Config", testConfig)
	suite.Run(t)
}

func testConfig(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	context("when a valid config is provided", func() {
		for _, tCase := range []struct {
			name     string
			before   MockEnv
			expected MockEnv
			config   string
		}{
			{
				"default is set when it doesn't exist",
				MockEnv{},
				MockEnv{
					"test": "test",
				},
				`api = "0.1"
				[[build.env]]
				name = "test"
				value = "test"`,
			},
			{
				"config is empty",
				MockEnv{},
				MockEnv{},
				`api = "0.1"`,
			},
			{
				"default is not set when it exists",
				MockEnv{
					"test": "test",
				},
				MockEnv{
					"test": "test",
				},
				`api = "0.1"
				[[build.env]]
				name = "test"
				value = "not-test"`,
			},
			{
				"override is set when it exists",
				MockEnv{
					"test": "test",
				},
				MockEnv{
					"test": "not-test",
				},
				`api = "0.1"
				[[build.env]]
				name = "test"
				value = "not-test"
				mode = "override"`,
			},
			{
				"env var is unset",
				MockEnv{
					"test": "test",
				},
				MockEnv{},
				`api = "0.1"
				[[build.env]]
				name = "test"
				mode = "unset"`,
			},
			{
				"env var is prepended with default separator",
				MockEnv{
					"test": "test",
				},
				MockEnv{
					"test": fmt.Sprintf("another-test%stest", string(os.PathListSeparator)),
				},
				`api = "0.1"
				[[build.env]]
				name = "test"
				value = "another-test"
				mode = "prepend"`,
			},
			{
				"env var is prepended with provided separator",
				MockEnv{
					"test": "test",
				},
				MockEnv{
					"test": "another-test$test",
				},
				`api = "0.1"
				[[build.env]]
				name = "test"
				value = "another-test"
				delim = "$"
				mode = "prepend"`,
			},
			{
				"env var is appended with default separator",
				MockEnv{
					"test": "another-test",
				},
				MockEnv{
					"test": fmt.Sprintf("another-test%stest", string(os.PathListSeparator)),
				},
				`api = "0.1"
				[[build.env]]
				name = "test"
				value = "test"
				mode = "append"`,
			},
			{
				"env var is appended with provided separator",
				MockEnv{
					"test": "another-test",
				},
				MockEnv{
					"test": "another-test$test",
				},
				`api = "0.1"
				[[build.env]]
				name = "test"
				value = "test"
				delim = "$"
				mode = "append"`,
			},
		} {
			tc := tCase
			it(tc.name, func() {
				tempDir := t.TempDir()
				configPath := path.Join(tempDir, "config.toml")
				Expect(os.WriteFile(configPath, []byte(tc.config), 0600)).To(BeNil())
				config := builderconfig.New(
					builderconfig.WithPath(configPath),
					builderconfig.WithEnvLooker(tc.before.Lookup),
					builderconfig.WithEnvSetter(tc.before.Set),
					builderconfig.WithEnvUnSetter(tc.before.UnSet),
				)
				Expect(config.Resolve()).To(BeNil())
				Expect(tc.before).To(Equal(tc.expected))
			})
		}
	})

	context("when an invalid config is provided", func() {
		for _, tCase := range []struct {
			name   string
			config string
			err    string
		}{
			{
				"invalid mode is provided",
				`api = "0.1"
				[[build.env]]
				name = "test"
				value = "test"
				delim = "$"
				mode = "unknown"`,
				"unknown mode",
			},
			{
				"invalid api is provided",
				`api = "01"`,
				"invalid API for builder config",
			},
			{
				"invalid toml is provided",
				`api = asd21312301"`,
				"unable to parse builder config",
			},
		} {
			tc := tCase
			it(tc.name, func() {
				tempDir := t.TempDir()
				configPath := path.Join(tempDir, "config.toml")
				Expect(os.WriteFile(configPath, []byte(tc.config), 0600)).To(BeNil())
				mock := MockEnv{}
				config := builderconfig.New(
					builderconfig.WithPath(configPath),
					builderconfig.WithEnvLooker(mock.Lookup),
					builderconfig.WithEnvSetter(mock.Set),
					builderconfig.WithEnvUnSetter(mock.UnSet),
				)
				err := config.Resolve()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring(tc.err))
			})

		}
	})

}

func (e MockEnv) Set(key string, value string) error {
	e[key] = value
	return nil
}

func (e MockEnv) UnSet(key string) error {
	delete(e, key)
	return nil
}

func (e MockEnv) Lookup(key string) (value string, exists bool) {
	value, exists = e[key]
	return
}
