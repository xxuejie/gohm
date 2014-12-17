package gohm

import (
	"testing"
)

type model struct {
	ID string `ohm:"id"`
	Name string `ohm:"name"`
}

func TestFetchTypeNameFromReturnInterface(t *testing.T) {
	name := fetchTypeNameFromReturnInterface(model{})
	if name != "model" {
		t.Errorf(`Expected type name: "model", actual type name: "%v"`, name)
	}
	name = fetchTypeNameFromReturnInterface(&model{})
	if name != "model" {
		t.Errorf(`Expected type name: "model", actual type name: "%v"`, name)
	}
	name = fetchTypeNameFromReturnInterface([]model{})
	if name != "model" {
		t.Errorf(`Expected type name: "model", actual type name: "%v"`, name)
	}
	name = fetchTypeNameFromReturnInterface(&[]model{})
	if name != "model" {
		t.Errorf(`Expected type name: "model", actual type name: "%v"`, name)
	}
}
