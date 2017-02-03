package main

import (
	"github.com/russross/blackfriday"
	"html/template"
	//"github.com/microcosm-cc/bluemonday"
	"bytes"
	"fmt"
	"github.com/fatih/structs"
	"io"
	"regexp"
)

type TemplateRenderer struct {
	Root string
	Skin string
}

func NewTemplateRenderer(tmplDir string, skin string) *TemplateRenderer {
	return &TemplateRenderer{Root: tmplDir, Skin: skin}
}

func (r *TemplateRenderer) renderTemplate(w io.Writer, tmpl string, wiki *Wiki, web string, p *Page) error {
	m := structs.Map(p)
	m["Web"] = web
	m["Webs"] = wiki.Webs

	templates := template.Must(template.New(r.Skin).
		Funcs(template.FuncMap{"md": createMarkdownRendering(m)}).ParseGlob(r.Root + "/" + r.Skin + "/*.html"))

	return templates.ExecuteTemplate(w, tmpl+".html", m)
}

// http://stackoverflow.com/questions/815787/what-perl-regex-can-match-camelcase-words
var wikiLinkMatcher = regexp.MustCompile(`(!)?\b([A-Z][a-z]+)?\.?([A-Z][a-zA-Z]*(?:[a-z][a-zA-Z]*[A-Z]|[A-Z][a-zA-Z]*[a-z])[a-zA-Z]*)\b`)

func createMarkdownRendering(m map[string]interface{}) func(...interface{}) template.HTML {
	return func(args ...interface{}) template.HTML {
		output := new(bytes.Buffer)
		tmpl, _ := template.New("_").Parse(fmt.Sprintf("%s", args...))
		tmpl.Execute(output, m)
		parsed := wikiLinkMatcher.ReplaceAllFunc(output.Bytes(), wikiLinkReplacer)
		unsafe := blackfriday.MarkdownCommon(parsed)
		//html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
		return template.HTML(unsafe)
	}
}

func wikiLinkReplacer(in []byte) []byte {
	m := wikiLinkMatcher.FindSubmatch(in)
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
