package provider

import (
	"reflect"
	"testing"
)

func getTeamStateDataV0() map[string]interface{} {
	return map[string]interface{}{
		"team_name": "foo123",
		"owners.#": "3",
		"owners.0": "bar",
		"owners.1": "baz",
		"owners.2": "qux",
		"pipeline_operators.#": "2",
		"pipeline_operators.0": "abc",
		"pipeline_operators.1": "def",
	}
}

func getTeamStateDataV1() map[string]interface{} {
	return map[string]interface{}{
		"team_name": "foo123",
		"owners.#": "3",
		"owners.1996459178": "bar",
		"owners.2015626392": "baz",
		"owners.2800005064": "qux",
		"pipeline_operators.#": "2",
		"pipeline_operators.891568578": "abc",
		"pipeline_operators.214229345": "def",
	}
}

func TestTeamStateUpgradeV0(t *testing.T) {
	expected := getTeamStateDataV1()
	actual, err := resourceTeamStateUpgradeV0(nil, getTeamStateDataV0(), nil)

	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}
