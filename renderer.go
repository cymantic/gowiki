package main

import (
	"html/template"
	"github.com/russross/blackfriday"
	//"github.com/microcosm-cc/bluemonday"
	"github.com/fatih/structs"
	"io"
	"fmt"
	"regexp"
	"bytes"
)

type Renderer struct {
	Root      string
	Skin      string
}

// http://stackoverflow.com/questions/815787/what-perl-regex-can-match-camelcase-words
var wikiLinkFinder = regexp.MustCompile(`(!)?\b([A-Z][a-z]+)?\.?([A-Z][a-zA-Z]*(?:[a-z][a-zA-Z]*[A-Z]|[A-Z][a-zA-Z]*[a-z])[a-zA-Z]*)\b`)

func wikiLinkReplacer(in []byte) []byte {
	m := wikiLinkFinder.FindSubmatch(in)
	if m != nil && len(m) == 4 {
		if m[1] != nil {
			if m[2] == nil {
				return m[3]
			} else {
				return []byte(string(m[2]) + "." + string(m[3]))
			}
		} else if m[2] == nil {
			return []byte("[" + string(m[3]) + "](" + string(m[3]) + ")")
		} else {
			return []byte("[" + string(m[2]) + "." + string(m[3]) + "](../" + string(m[2]) + "/" + string(m[3]) + ")")
		}
	}
	link := string(in)
	return []byte("[" + link + "](" + link + ")")
}

func createWikiRenderer(m map[string]interface{}) func(...interface{}) template.HTML {
	return func(args ...interface{}) template.HTML {
		output := new(bytes.Buffer)
		tmpl, _ := template.New("_").Parse(fmt.Sprintf("%s", args...))
		tmpl.Execute(output, m)
		parsed := wikiLinkFinder.ReplaceAllFunc(output.Bytes(), wikiLinkReplacer)
		unsafe := blackfriday.MarkdownCommon(parsed)
		//html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
		return template.HTML(unsafe)
	}
}

func NewRenderer(tmplDir string, skin string) *Renderer {
	return &Renderer{Root:tmplDir,Skin:skin}
}

func (r *Renderer) renderTemplate(w io.Writer, tmpl string, web string, p *Page) error {
	m := structs.Map(p)
	m["Web"] = web

	templates := template.Must(template.New(r.Skin).
		Funcs(template.FuncMap{"md": createWikiRenderer(m)}).ParseGlob(r.Root + "/" + r.Skin + "/*.html"))

	return templates.ExecuteTemplate(w, tmpl+".html", m)
}
