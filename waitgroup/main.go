package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

const nPages = 100
const nParsers = 20

type pageLinks struct {
	Page  int
	Links []*pageLinks
	Seen  bool
}

type pageParser struct {
	pages    chan *pageLinks
	wg       sync.WaitGroup
	queueLen atomic.Int32
}

func makeParser() *pageParser {
	parser := new(pageParser)
	parser.pages = make(chan *pageLinks, 1)
	parser.wg = sync.WaitGroup{}
	return parser
}

func crawl(rootPage *pageLinks) {
	parser := makeParser()
	parser.sendPage(rootPage)
	for i := 0; i < nParsers; i++ {
		parser.wg.Add(1)
		go parser.parse()
	}
	parser.wg.Wait()
	close(parser.pages)
}

func (p *pageParser) sendPage(toParse *pageLinks) {
	if !toParse.Seen {
		p.queueLen.Add(1)
		// Send on a separate goroutine so we don't block
		go func() {
			p.pages <- toParse
		}()
	}
}

func (p *pageParser) parse() {
	defer p.wg.Done()
	for p.queueLen.Load() > 0 {
		select {
		case page := <-p.pages:
			if page.Seen {
				p.queueLen.Add(-1)
				break
			}
			page.Seen = true

			fmt.Printf("Parsing %v\n", page.Page)
			time.Sleep(time.Millisecond * 250)
			for _, page := range page.Links {
				p.sendPage(page)
			}
			p.queueLen.Add(-1)
		default:
		}
	}
}

func main() {
	pages := make([]pageLinks, nPages)

	for i := 0; i < nPages; i++ {
		nLinks := rand.Intn(3) + 1
		pages[i].Page = i
		pages[i].Links = make([]*pageLinks, nLinks)

		for n := 0; n < nLinks; n++ {
			pages[i].Links[n] = &pages[rand.Intn(nPages)]
		}
	}

	crawl(&pages[0])

	verify(pages)
}

func verify(pages []pageLinks) {
	crawled := make([]bool, nPages)
	listCrawled(&pages[0], crawled)
	for i := 0; i < nPages; i++ {
		if crawled[i] != pages[i].Seen {
			fmt.Printf("Page %v, expected %v, actual %v\n", i, crawled[i], pages[i].Seen)
		}
	}
}

func listCrawled(page *pageLinks, crawled []bool) {
	crawled[page.Page] = true

	for _, p := range page.Links {
		if !crawled[p.Page] {
			listCrawled(p, crawled)
		}
	}
}
