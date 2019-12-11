package matchers

import (
	"fmt"
	"reflect"

	"github.com/BurntSushi/toml"
	"github.com/onsi/gomega/types"
)

func MatchTOML(expected interface{}) types.GomegaMatcher {
	return &matchTOML{
		expected: expected,
	}
}

type matchTOML struct {
	expected interface{}
}

func (matcher *matchTOML) Match(actual interface{}) (success bool, err error) {
	var e, a string

	switch eType := matcher.expected.(type) {
	case string:
		e = eType
	case []byte:
		e = string(eType)
	default:
		return false, fmt.Errorf("expected value must be []byte or string, received %T", matcher.expected)
	}

	switch aType := actual.(type) {
	case string:
		a = aType
	case []byte:
		a = string(aType)
	default:
		return false, fmt.Errorf("actual value must be []byte or string, received %T", matcher.expected)
	}

	var eValue map[string]interface{}
	_, err = toml.Decode(e, &eValue)
	if err != nil {
		return false, err
	}

	var aValue map[string]interface{}
	_, err = toml.Decode(a, &aValue)
	if err != nil {
		return false, err
	}

	return reflect.DeepEqual(eValue, aValue), nil
}

func (matcher *matchTOML) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n%s\nto match the TOML representation of\n%s", actual, matcher.expected)
}

func (matcher *matchTOML) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n%s\nnot to match the TOML representation of\n%s", actual, matcher.expected)
}
