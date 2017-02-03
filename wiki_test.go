package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseWebAndTitleFromURL(t *testing.T) {
	web, title, err := parseTitleFromURL("/view/Main/WebTest")
	if err != nil {
		t.Errorf("can't parse URL %s", "/view/Main/WebTest")
	}

	if title != "WebTest" {
		t.Errorf("expected '%s' got '%s'", "WebTest", title)
	}

	if web != "Main" {
		t.Errorf("expected '%s' got '%s'", "Main", web)
	}
}

func makeViewRequest(wikiRepository WikiRepository) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", "/view/Main/WebPage", nil)
	rr := httptest.NewRecorder()
	renderer := NewTemplateRenderer("tmpl", "default")
	handler := http.HandlerFunc(makeHandler(viewHandler, wikiRepository, renderer))
	handler.ServeHTTP(rr, req)
	return rr
}

func TestViewFound(t *testing.T) {
	rr := makeViewRequest(fakeWikiRepositoryWithFile)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected '%s' got '%s'", http.StatusOK, status)
	}
}

func TestViewNotFound(t *testing.T) {
	rr := makeViewRequest(fakeWikiRepositoryNoFile)
	if status := rr.Code; status != http.StatusFound {
		t.Errorf("expected %v got %v", http.StatusFound, status)
	}

	if moved := rr.HeaderMap.Get("Location"); moved != "/edit/Main/WebPage" {
		t.Errorf("expected '%s' got '%s'", "/edit/Main/WebPage", moved)
	}
}
