package main

import (
	"net/http"
	"testing"
	"net/http/httptest"
	"errors"
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

type fakePageStorageFound struct {
}

func (t *fakePageStorageFound) WritePage(web string, p *Page) error {
	return nil
}

func (t *fakePageStorageFound) ReadPage(web string, title string) (*Page, error) {
	return &Page{Title:title, Body:[]byte("Hello, world!")}, nil
}

func (t *fakePageStorageFound) CreateWeb(web string) error {
	return nil
}


type fakePageStorageNotFound struct {
}

func (t *fakePageStorageNotFound) WritePage(web string, p *Page) error {
	return nil
}

func (t *fakePageStorageNotFound) ReadPage(web string, title string) (*Page, error) {
	return nil, errors.New("no file")
}

func (t *fakePageStorageNotFound) CreateWeb(web string) error {
	return nil
}

func makeViewRequest(fakeStorage PageStorage) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", "/view/Main/WebPage", nil)
	rr := httptest.NewRecorder()
	renderer := NewRenderer("default")
	handler := http.HandlerFunc(makeHandler(viewHandler, fakeStorage, renderer))
	handler.ServeHTTP(rr, req)
	return rr
}

func TestViewFound(t *testing.T) {
	rr := makeViewRequest(&fakePageStorageFound{})
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected '%s' got '%s'", http.StatusOK, status)
	}
}

func TestViewNotFound(t *testing.T) {
	rr := makeViewRequest(&fakePageStorageNotFound{})
	if status := rr.Code; status != http.StatusFound {
		t.Errorf("expected %v got %v", http.StatusFound, status)
	}

	if moved := rr.HeaderMap.Get("Location"); moved != "/edit/Main/WebPage" {
		t.Errorf("expected '%s' got '%s'", "/edit/Main/WebPage", moved)
	}
}
