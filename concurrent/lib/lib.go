package lib

const MaxParsers = 20

type Crawler interface {
	Crawl()
}

type PageCrawler struct {
	RootPage *PageLinks
}

type PageLinks struct {
	Page  int
	Links []*PageLinks
	Seen  bool
}
