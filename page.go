package main

type Page struct {
	Title string
	Body  []byte
}

type PageStorage interface {
	WritePage(web string, p *Page) error
	ReadPage(web string, title string) (*Page, error)
	CreateWeb(web string) error
}

func (p *Page) save(web string, s PageStorage) error {
	return s.WritePage(web, p)
}

func loadPage(s PageStorage, web string, title string) (*Page, error) {
	return s.ReadPage(web, title)
}

