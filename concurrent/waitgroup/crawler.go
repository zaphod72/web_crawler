package waitgroup

import (
	"fmt"
	"sync"
	"time"

	"github.com/zaphod72/web_crawler/concurrent/lib"
)

type pageLinks *lib.PageLinks

type Crawler struct {
	lib.PageCrawler
}

func (c Crawler) Crawl() {
	wg := sync.WaitGroup{}
	pages := make(chan pageLinks)
	done := make(chan struct{})

	for i := 0; i < lib.MaxParsers; i++ {
		wg.Add(1)
		go parse(&wg, pages, done)
	}
	pages <- c.RootPage
	wg.Wait()
	close(done)
}

func parse(
	wg *sync.WaitGroup,
	pages chan pageLinks,
	done <-chan struct{}) {

	for {
		select {
		case page := <-pages:
			wg.Add(parsePage(page, pages))
		case <-done:
			return
		}
	}
}

// parsePage will queue all not seen links found in the page
// as long as page is not seen
// Returns the change in the queue size, i.e. -1 for this page parsed +1 for each not seen link
func parsePage(page pageLinks, pages chan<- pageLinks) int {
	queueLenChange := -1

	// Read and write of page seen - make it atomic
	seen := page.AtomicSeen.Swap(true)
	if seen {
		return queueLenChange
	}

	page.Seen = true // Just so the verify function works

	// Fetching page
	fmt.Printf("Loading %v\n", page.Page)
	time.Sleep(time.Millisecond * 250)

	// Queue links
	for _, link := range page.Links {
		if link.Seen {
			continue
		}
		link := link
		go func() {
			pages <- link
		}()
		queueLenChange++
	}
	return queueLenChange
}
