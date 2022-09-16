package worker_pool

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/zaphod72/web_crawler/concurrent/lib"
)

type twoChannelParser struct {
	toLoad   chan pageLinks
	toParse  chan pageLinks
	queueLen atomic.Int32
	done     chan struct{}
}

type TwoChannelCrawler struct {
	lib.PageCrawler
}

func NewTwoChannelParser() *twoChannelParser {
	parser := twoChannelParser{
		toLoad:  make(chan pageLinks),
		toParse: make(chan pageLinks),
		done:    make(chan struct{}),
	}
	return &parser
}

func (c TwoChannelCrawler) Crawl() {
	parser := NewTwoChannelParser()
	for i := 0; i < 1; i++ {
		go parser.load()
	}
	parser.toLoad <- c.RootPage
	parser.queueLen.Add(1)
	// parse() returns when there is no work left
	parser.parse()
	/*
		for i := 0; i < lib.MaxParsers; i++ {
			parser.done <- struct{}{}
		}
	*/
	close(parser.done)
}

func (p *twoChannelParser) load() {
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
			return
		}
	}
}

func (p *twoChannelParser) parse() {
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
					p.toLoad <- link
				}
			}()
		}
		p.queueLen.Add(-1)
	}
}
