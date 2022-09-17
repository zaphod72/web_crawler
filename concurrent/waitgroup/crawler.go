package waitgroup

import (
	"fmt"
	"sync"
	"sync/atomic"
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
	workCounter := make(chan int)
	done := make(chan struct{})

	for i := 0; i < lib.MaxParsers; i++ {
		wg.Add(1)
		go parse(&wg, pages, workCounter, done)
	}
	pages <- c.RootPage
	monitorWorkCounter(workCounter)
	close(done)
	wg.Wait()
}

func monitorWorkCounter(workCounter <-chan int) {
	workCount := atomic.Int32{}
	for {
		select {
		case queueChange := <-workCounter:
			if workCount.Add(int32(queueChange)) == 0 {
				return
			}
		}
	}
}

func parse(wg *sync.WaitGroup, pages chan pageLinks, workCounter chan<- int, done <-chan struct{}) {
	defer wg.Done()
	for {
		select {
		case page := <-pages:
			workCounter <- parsePage(page, pages)
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
	if page.Seen {
		return queueLenChange
	}
	// Fetching page
	fmt.Printf("Loading %v\n", page.Page)
	time.Sleep(time.Millisecond * 250)
	page.Seen = true

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
