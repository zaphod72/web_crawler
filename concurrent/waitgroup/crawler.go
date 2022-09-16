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
	toLoad   chan pageLinks
	toParse  chan pageLinks
	done     chan struct{}
	wg       sync.WaitGroup
	queueLen atomic.Int32
}

type Crawler struct {
	lib.PageCrawler
}

func NewParser() *pageParser {
	return &pageParser{
		toLoad:  make(chan pageLinks),
		toParse: make(chan pageLinks),
		wg:      sync.WaitGroup{},
	}
}

func (c Crawler) Crawl() {
	parser := NewParser()
	for i := 0; i < lib.MaxParsers; i++ {
		parser.wg.Add(1)
		go parser.load()
	}
	parser.toLoad <- c.RootPage
	parser.queueLen.Add(1)
	// parse() returns when there is no work left
	parser.parse()
	// Closing the toLoad channel here would send nil to page := <-p.toLoad
	// and a NPE on page.Seen
	// Using the done channel instead avoids this.
	parser.done <- struct{}{}
	parser.wg.Wait()
}

func (p *pageParser) load() {
	defer p.wg.Done()
	for {
		select {
		case page := <-p.toLoad:
			if page.Seen {
				p.queueLen.Add(-1)
				continue
			}
			// Fetching page
			fmt.Printf("Loading %v\n", page.Page)
			time.Sleep(time.Millisecond * 250)
			page.Seen = true
			p.toParse <- page
		case <-p.done:
		}
	}
}

func (p *pageParser) parse() {
	for p.queueLen.Load() > 0 {
		page := <-p.toParse
		fmt.Printf("Parsing %v\n", page.Page)
		for _, link := range page.Links {
			if link.Seen {
				continue
			}
			p.queueLen.Add(1)
			link := link
			go func() {
				if link.Seen {
					p.queueLen.Add(-1)
				} else {
					go func() {
						p.toLoad <- link
					}()
				}
			}()
		}
		p.queueLen.Add(-1)
	}
}
