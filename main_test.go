package main

import "testing"

func TestHasFlag(t *testing.T) {
	if !hasFlag([]string{"--json"}, "--json") {
		t.Error("expected true for ['--json'], '--json'")
	}
	if !hasFlag([]string{"somename", "--json"}, "--json") {
		t.Error("expected true when --json is among other args")
	}
	if hasFlag([]string{"somename"}, "--json") {
		t.Error("expected false when --json is absent")
	}
	if hasFlag([]string{}, "--json") {
		t.Error("expected false for empty args")
	}
}
