package main

type Page struct {
	Title string
	Body  []byte
}

type PageStorage interface {
	WritePage(p *Page) error
	ReadPage(title string) (*Page, error)
}

func (p *Page) save(s PageStorage) error {
	return s.WritePage(p)
}

func loadPage(s PageStorage, title string) (*Page, error) {
	return s.ReadPage(title)
}

