package main

import (
	"html/template"
	"github.com/russross/blackfriday"
	"io"
	"fmt"
)

type Renderer struct {
	Skin      string
	Templates *template.Template
}

func markDown(args ...interface{}) template.HTML {
	s := blackfriday.MarkdownCommon([]byte(fmt.Sprintf("%s", args...)))
	return template.HTML(s)
}

func NewRenderer(skin string) *Renderer {
	templates := template.Must(
		template.New(skin).Funcs(template.FuncMap{"md": markDown}).ParseGlob("tmpl/" + skin + "/*.html"))
	return &Renderer{Skin:skin, Templates:templates}
}

func (r *Renderer) renderTemplate(w io.Writer, tmpl string, p *Page) error {
	return r.Templates.ExecuteTemplate(w, tmpl+".html", p)
}
