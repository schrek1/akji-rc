package main

import "testing"

func TestProbeMessage(t *testing.T) {
	if probeMessage == "" {
		t.Fatal("probe message must not be empty")
	}
}
