package main

import "testing"

func TestGenerateWikiLinks(t *testing.T) {
	validateGenerateWikiLinks(t, "!WikiLink", "WikiLink")
	validateGenerateWikiLinks(t, "WikiLink", "[WikiLink](WikiLink)")
	validateGenerateWikiLinks(t, "Web.WikiLink", "[Web.WikiLink](../Web/WikiLink)")
	validateGenerateWikiLinks(t, "!Web.WikiLink", "Web.WikiLink")
}
func validateGenerateWikiLinks(t *testing.T, input string, expected string) {
	in := []byte(input)
	output := string(wikiLinkReplacer(in))
	if output != expected {
		t.Errorf("expected '%s' got '%s'", expected, output)
	}
}

func TestMatchAndReplaceWikiLinks(t *testing.T) {
	validateMatchAndReplace(t, "some WikiLink here", "some [WikiLink](WikiLink) here")
	validateMatchAndReplace(t, "some Web.WikiLink here", "some [Web.WikiLink](../Web/WikiLink) here")
}
func validateMatchAndReplace(t *testing.T, input string, expected string) {
	in := []byte(input)
	output := string(wikiLinkFinder.ReplaceAllFunc(in, wikiLinkReplacer))
	if output != expected {
		t.Errorf("expected '%s' got '%s'", expected, output);
	}
}
