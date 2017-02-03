package main

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/bmizerany/pat"
	"net/http"
	"regexp"
)

type Page struct {
	Title string
	Body  []byte
	Meta  map[string]interface{}
}

type Web struct {
	Name string
	Settings map[string]interface{}
}

type Wiki struct {
	Repository   WikiRepository
	PageRenderer *TemplateRenderer
	Webs         map[string]*Web
}

type WikiRepository interface {
	CreateWeb(web string) (*Web, error)
	LoadWebs() map[string]*Web
	WritePage(web string, p *Page) error
	ReadPage(web string, title string) (*Page, error)
}

func NewWiki(wikiRepository WikiRepository, templateRenderer *TemplateRenderer) *Wiki {
	webs:= wikiRepository.LoadWebs()
	wiki := &Wiki{Repository: wikiRepository, PageRenderer: templateRenderer, Webs: webs}
	configureHTTPHandlers(wiki, wikiRepository, templateRenderer)
	return wiki
}

func configureHTTPHandlers(wiki *Wiki, wikiRepository WikiRepository, pageRenderer *TemplateRenderer) {
	m := pat.New()
	m.Get("/", http.HandlerFunc(routeToMainWebHomeHandler))
	m.Get("/view", http.HandlerFunc(routeToMainWebHomeHandler))
	m.Get("/view/:web", http.HandlerFunc(routeToWebHomeHandler))
	m.Get("/view/:web/:title", makeHandler(viewHandler, wiki, wikiRepository, pageRenderer))
	m.Get("/edit/:web/:title", makeHandler(editHandler, wiki, wikiRepository, pageRenderer))
	m.Post("/save/:web/:title", makeSaveHandler(saveHandler, wiki, wikiRepository))
	m.Post("/web/:web/:title", makeSaveHandler(createWebHandler, wiki, wikiRepository))
	http.Handle("/", m)
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

func renderTemplate(w http.ResponseWriter, r *TemplateRenderer, tmpl string, wiki *Wiki, web string, p *Page) {
	err := r.renderTemplate(w, tmpl, wiki, web, p)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, wiki *Wiki, wikiRepository WikiRepository, templateRenderer *TemplateRenderer, web string, title string) {
	p, err := loadPage(wikiRepository, web, title)
	if err != nil {
		http.Redirect(w, r, generatePath("edit", web, title), http.StatusFound)
		return
	}
	renderTemplate(w, templateRenderer, "view", wiki, web, p)
}

func editHandler(w http.ResponseWriter, r *http.Request, wiki *Wiki,
	wikiRepository WikiRepository, templateRenderer *TemplateRenderer,
	web string, title string) {
	p, err := loadPage(wikiRepository, web, title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, templateRenderer, "edit", wiki, web, p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, wiki *Wiki, wikiRepository WikiRepository, web string, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save(wikiRepository, web)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, generatePath("view", web, title), http.StatusFound)
}

var validWeb = regexp.MustCompile(`^[A-Z][a-z]+$`)

func createWebHandler(w http.ResponseWriter, r *http.Request, wiki *Wiki, wikiRepository WikiRepository, web string, title string) {
	name := r.FormValue("name")
	if !validWeb.MatchString(name) {
		log.Error("Bad Web Name " + name)
		http.Error(w, "Bad Web Name "+name, http.StatusBadRequest)
		return
	}
	webDefinition, err := wikiRepository.CreateWeb(name)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	wiki.Webs[webDefinition.Name] = webDefinition
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

func makeHandler(fn func(http.ResponseWriter, *http.Request, *Wiki, WikiRepository, *TemplateRenderer, string, string),
	wiki *Wiki, wikiRepository WikiRepository, templateRenderer *TemplateRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		web, title, err := parseTitleFromURL(r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, wiki, wikiRepository, templateRenderer, web, title)
	}
}

func makeSaveHandler(fn func(http.ResponseWriter, *http.Request, *Wiki, WikiRepository, string, string),
	wiki *Wiki, wikiRepository WikiRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		web, title, err := parseTitleFromURL(r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, wiki, wikiRepository, web, title)
	}
}
