package main

func (p *Page) save(wikiRepository WikiRepository, web string) error {
	return wikiRepository.WritePage(web, p)
}

func loadPage(wikiRepository WikiRepository, web string, title string) (*Page, error) {
	return wikiRepository.ReadPage(web, title)
}
