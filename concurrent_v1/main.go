package main

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"
)

const nPages = 100
const maxParsers = 20

type pageLinks struct {
	Page  int
	Links []*pageLinks
	Seen  bool
}

type pageParser struct {
	pages    chan *pageLinks
	count    atomic.Int32
	queueLen atomic.Int32
	done     chan bool
}

func makeParser() *pageParser {
	parser := new(pageParser)
	parser.pages = make(chan *pageLinks)
	parser.done = make(chan bool)
	return parser
}

func crawl(rootPage *pageLinks) {
	parser := makeParser()
	parser.sendPage(rootPage)
	parser.startParser()
	<-parser.done
	close(parser.done)
	close(parser.pages)
}

func (p *pageParser) sendPage(toParse *pageLinks) bool {
	if !toParse.Seen {
		p.queueLen.Add(1)
		// Send on a separate goroutine so we don't block
		go func() {
			p.pages <- toParse
		}()
		return true
	}
	return false
}

func (p *pageParser) startParser() {
	if p.count.Load() < maxParsers {
		// Don't know when the goroutine will start so increment parser count now
		count := p.count.Add(1)
		fmt.Printf("Started parser %v\n", count)
		go p.parse()
	}
}

func (p *pageParser) parse() {
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
			for i, page := range page.Links {
				// Not starting another parse goroutine for the first link as this goroutine can parse it
				if p.sendPage(page) && i > 0 {
					p.startParser()
				}
			}
			p.queueLen.Add(-1)
		default:
		}
	}

	if p.count.Add(-1) == 0 {
		p.done <- true
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
