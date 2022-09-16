package waitgroup

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zaphod72/web_crawler/concurrent/lib"
)

type pageLinks *lib.PageLinks

type pageParser struct {
	pages    chan pageLinks
	wg       sync.WaitGroup
	queueLen atomic.Int32
}

type Crawler struct {
	lib.PageCrawler
}

func makeParser() *pageParser {
	parser := new(pageParser)
	parser.pages = make(chan pageLinks, 1)
	parser.wg = sync.WaitGroup{}
	return parser
}

func (c Crawler) Crawl() {
	parser := makeParser()
	parser.sendPage(c.RootPage)
	for i := 0; i < lib.MaxParsers; i++ {
		parser.wg.Add(1)
		go parser.parse()
	}
	parser.wg.Wait()
	close(parser.pages)
}

func (p *pageParser) sendPage(toParse pageLinks) {
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
