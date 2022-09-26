package waitgroup

import (
	"sync"
	"sync/atomic"

	"github.com/zaphod72/web_crawler/concurrent/lib"
)

func (c Crawler) CrawlWithCounter() {
	wg := sync.WaitGroup{}
	pages := make(chan pageLinks)
	workCounter := make(chan int)
	done := make(chan struct{})

	for i := 0; i < lib.MaxParsers; i++ {
		wg.Add(1)
		go parseWithCounter(&wg, pages, workCounter, done)
	}
	pages <- c.RootPage
	monitorWorkCounter(workCounter, 1)
	close(done)
	wg.Wait()
}

func monitorWorkCounter(workCounter <-chan int, initialValue int32) {
	workCount := atomic.Int32{}
	workCount.Store(initialValue)
	for {
		select {
		case queueChange := <-workCounter:
			if workCount.Add(int32(queueChange)) == 0 {
				return
			}
		}
	}
}

func parseWithCounter(
	wg *sync.WaitGroup,
	pages chan pageLinks,
	workCounter chan<- int,
	done <-chan struct{}) {

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
