package main

import (
	"net/http"
	"regexp"
	"errors"
)

type Wiki struct {
	Renderer *Renderer
	Storage PageStorage
}

func NewWiki(renderer *Renderer, storage PageStorage) *Wiki {
	http.HandleFunc("/view/", makeHandler(viewHandler, storage, renderer))
	http.HandleFunc("/edit/", makeHandler(editHandler, storage, renderer))
	http.HandleFunc("/save/", makeHandler(saveHandler, storage, renderer))
	return &Wiki{Renderer:renderer, Storage:storage}
}

func (w *Wiki) start(address string) {
	http.ListenAndServe(address, nil)
}

func renderTemplate(w http.ResponseWriter, r *Renderer, tmpl string, p *Page) {
	err := r.renderTemplate(w, tmpl, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, s PageStorage, rend *Renderer, title string) {
	p, err := loadPage(s, title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, rend, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, s PageStorage, rend *Renderer, title string) {
	p, err := loadPage(s, title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, rend, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, s PageStorage, rend *Renderer, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save(s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func parseTitleFromURL(path string) (string, error) {
	m := validPath.FindStringSubmatch(path)
	if m == nil {
		return "", errors.New("Not Found")
	}
	return m[2], nil
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, PageStorage, *Renderer, string), s PageStorage, rend *Renderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title, err := parseTitleFromURL(r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, s, rend, title)
	}
}
