package worker_pool

import (
	"fmt"
	"time"

	"github.com/zaphod72/web_crawler/concurrent/lib"
)

func (c TwoChannelCrawler) _Crawl() {
	parser := NewTwoChannelParser()
	for i := 0; i < 1; i++ {
		go parser.load()
	}
	parser.toLoad <- c.RootPage
	parser.queueLen.Add(1)
	// parse() returns when there is no work left
	parser.parse()
	for i := 0; i < lib.MaxParsers; i++ {
		parser.done <- struct{}{}
	}
}

func (p *twoChannelParser) _load() {
	for {
		select {
		case page := <-p.toLoad:
			// Fetching page
			fmt.Printf("Loading %v\n", page.Page)
			time.Sleep(time.Millisecond * 250)
			page.Seen = true
			p.toParse <- page
		case <-p.done:
			return
		}
	}
}

func (p *twoChannelParser) _parse() {
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
				if !link.Seen {
					p.toLoad <- link
				}
			}()
		}
		p.queueLen.Add(-1)
	}
}
