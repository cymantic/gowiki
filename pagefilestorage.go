package main

import (
	"os"
	"io/ioutil"
)

type FilePageStorage struct {
	Root string
}

func NewFilePageStorage(path string) *FilePageStorage {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}
	return &FilePageStorage{Root:path}
}

func pageToFilename(root string, title string) string {
	return root + "/" + title + ".md"
}

func (s *FilePageStorage) WritePage(p *Page) error {
	filename := pageToFilename(s.Root, p.Title)
	return ioutil.WriteFile(filename, p.Body, 0644)
}

func (s *FilePageStorage) ReadPage(title string) (*Page, error) {
	filename := pageToFilename(s.Root, title)
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body:body}, nil
}
