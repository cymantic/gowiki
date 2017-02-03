package main

import (
	"testing"
	"errors"
)

func TestPageToFilename(t *testing.T) {
	filename := pageToFilename("data", "Main", "WebTest")
	if filename != "data/Main/WebTest.md" {
		t.Errorf("expected '%s' got '%s'", "WebTest.md", filename)
	}
}


type FakeWikiRepository struct {
	readFn func(string, string) (*Page, error)
}

func NewFakeWikiRepository(fn func(string, string) (*Page, error)) *FakeWikiRepository {
	return &FakeWikiRepository{readFn: fn}
}

func (f *FakeWikiRepository) WritePage(web string, p *Page) error {
	return nil
}

func (f *FakeWikiRepository) ReadPage(web string, title string) (*Page, error) {
	return f.readFn(web, title)
}

func (f *FakeWikiRepository) CreateWeb(web string) error {
	return nil
}

func (f *FakeWikiRepository) LoadWebs() map[string]*Web {
	return map[string]*Web {
		"Main":&Web{},
		"Sandbox":&Web{},
	}
}


var fakeWikiRepositoryWithFile = NewFakeWikiRepository(func(web string, title string) (*Page, error) {
	return &Page{Title: title, Body: []byte("Hello, world!")}, nil
})

var fakeWikiRepositoryNoFile = NewFakeWikiRepository(func(web string, title string) (*Page, error) {
	return nil, errors.New("file not found")
})
