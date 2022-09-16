package worker_pool

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/zaphod72/web_crawler/concurrent/lib"
)

type pageLinks *lib.PageLinks

type pageParser struct {
	pages    chan pageLinks
	count    atomic.Int32
	queueLen atomic.Int32
	done     chan bool
}

type Crawler struct {
	lib.PageCrawler
}

func makeParser() *pageParser {
	parser := new(pageParser)
	parser.pages = make(chan pageLinks)
	parser.done = make(chan bool)
	return parser
}

func (c Crawler) Crawl() {
	parser := makeParser()
	parser.sendPage(c.PageCrawler.RootPage)
	parser.startParser()
	<-parser.done
	close(parser.done)
	close(parser.pages)
}

func (p *pageParser) sendPage(toParse pageLinks) bool {
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
	if p.count.Load() < lib.MaxParsers {
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
