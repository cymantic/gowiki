package main

import "testing"

func TestPageToFilename(t *testing.T) {
	filename := pageToFilename("data","Main", "WebTest")
	if filename != "data/Main/WebTest.md" {
		t.Errorf("expected '%s' got '%s'", "WebTest.md", filename)
	}
}
