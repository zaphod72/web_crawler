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
	pages chan *pageLinks
	count atomic.Int32
	done  chan bool
}

func makeParser() *pageParser {
	parser := new(pageParser)
	parser.pages = make(chan *pageLinks)
	parser.done = make(chan bool)
	return parser
}

func crawl(rootPage pageLinks) {
	parser := makeParser()
	parser.sendPage(&rootPage)
	parser.startParser()
	<-parser.done
	close(parser.done)
	close(parser.pages)
}

func (p *pageParser) sendPage(toParse *pageLinks) bool {
	if !toParse.Seen {
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
	for p.count.Load() > 0 {
		var page *pageLinks
		select {
		case page = <-p.pages:
			if page.Seen {
				continue
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
		default:
			p.close()
		}
	}
}

func (p *pageParser) close() {
	count := p.count.Add(-1)
	fmt.Printf("Close parser. Remaining %v\n", count)
	if count == 0 {
		p.done <- true
	}
}

func main() {
	pages := make([]pageLinks, nPages)

	for i := 0; i < nPages; i++ {
		nLinks := rand.Intn(15) + 5
		pages[i].Page = i
		pages[i].Links = make([]*pageLinks, nLinks)

		for n := 0; n < nLinks; n++ {
			pages[i].Links[n] = &pages[rand.Intn(nPages)]
		}
	}

	crawl(pages[0])
}
