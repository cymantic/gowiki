package main

import (
	"net/http"
	"testing"
	"net/http/httptest"
)

func TestPageToFilename(t *testing.T) {
	filename := pageToFilename("data","test")
	if filename != "data/test.md" {
		t.Errorf("expected %s got %s", "test.md", filename)
	}
}

func TestParseTitleFromURL(t *testing.T) {
	title, err := parseTitleFromURL("/view/test")
	if err != nil {
		t.Errorf("can't parse URL %s", "/view/test")
	}

	if title != "test" {
		t.Errorf("expected %s got %s", "test", title)
	}
}

type testPageStorage struct {
}

func (t *testPageStorage) WritePage(p *Page) error {
	return nil
}

func (t *testPageStorage) ReadPage(title string) (*Page, error) {
	return &Page{Title:title, Body:[]byte("Hello, world!")}, nil
}


func TestViewNotFound(t *testing.T) {
	req, _ := http.NewRequest("GET", "/view/test", nil)
	rr := httptest.NewRecorder()
	s := &testPageStorage{}
	rend := NewRenderer("default")

	handler := http.HandlerFunc(makeHandler(viewHandler, s, rend))

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected %v got %v", http.StatusOK, status)
	}
}
