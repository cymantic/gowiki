package main

import (
	"net/http"
	"github.com/bmizerany/pat"
	"regexp"
	"errors"
	log "github.com/Sirupsen/logrus"
)

type Wiki struct {
	Renderer *Renderer
	Storage  PageStorage
}

func NewWiki(renderer *Renderer, storage PageStorage) *Wiki {
	m := pat.New()

	m.Get("/", http.HandlerFunc(routeToMainWebHomeHandler))
	m.Get("/view", http.HandlerFunc(routeToMainWebHomeHandler))
	m.Get("/view/:web", http.HandlerFunc(routeToWebHomeHandler))
	m.Get("/view/:web/:title", makeHandler(viewHandler, storage, renderer))
	m.Get("/edit/:web/:title", makeHandler(editHandler, storage, renderer))
	m.Post("/save/:web/:title", makeSaveHandler(saveHandler, storage))
	m.Post("/web/:web/:title", makeSaveHandler(createWebHandler, storage))

	http.Handle("/", m)
	return &Wiki{Renderer:renderer, Storage:storage}
}

func (w *Wiki) start(address string) error {
	return http.ListenAndServe(address, nil)
}

func generatePath(action string, web string, title string) string {
	return "/" + action + "/" + web + "/" + title
}

func routeToMainWebHomeHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, generatePath("view", "Main", "WebHome"), http.StatusFound)
	return
}

func routeToWebHomeHandler(w http.ResponseWriter, r *http.Request) {
	web := r.URL.Query().Get(":web")
	if web == "" {
		web = "Main"
	}
	http.Redirect(w, r, generatePath("view", web, "WebHome"), http.StatusFound)
	return
}

func renderTemplate(w http.ResponseWriter, r *Renderer, tmpl string, web string, p *Page) {
	err := r.renderTemplate(w, tmpl, web, p)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, s PageStorage, renderer *Renderer, web string, title string) {
	p, err := loadPage(s, web, title)
	if err != nil {
		http.Redirect(w, r, generatePath("edit", web, title), http.StatusFound)
		return
	}
	renderTemplate(w, renderer, "view", web, p)
}

func editHandler(w http.ResponseWriter, r *http.Request, s PageStorage, renderer *Renderer, web string, title string) {
	p, err := loadPage(s, web, title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, renderer, "edit", web, p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, s PageStorage, web string, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save(web, s)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, generatePath("view", web, title), http.StatusFound)
}

var validWeb = regexp.MustCompile(`^[A-Z][a-z]+$`)

func createWebHandler(w http.ResponseWriter, r *http.Request, s PageStorage, web string, title string) {
	name := r.FormValue("name")
	if !validWeb.MatchString(name) {
		log.Error("Bad Web Name " + name)
		http.Error(w, "Bad Web Name " + name, http.StatusBadRequest)
		return
	}
	err := s.CreateWeb(name)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, generatePath("view", name, "WebHome"), http.StatusFound)
}

var validPath = regexp.MustCompile(`^/(edit|save|view|web)/([a-zA-Z0-9]+)/([a-zA-Z0-9]+)$`)

func parseTitleFromURL(path string) (string, string, error) {
	m := validPath.FindStringSubmatch(path)
	if m == nil {
		return "", "", errors.New("Not Found")
	}
	return m[2], m[3], nil
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, PageStorage, *Renderer, string, string), s PageStorage, renderer *Renderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		web, title, err := parseTitleFromURL(r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, s, renderer, web, title)
	}
}

func makeSaveHandler(fn func(http.ResponseWriter, *http.Request, PageStorage, string, string), s PageStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		web, title, err := parseTitleFromURL(r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, s, web, title)
	}
}

