package fanout

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

func queuePages(done <-chan bool, links <-chan pageLinks) <-chan pageLinks {
	toParse := make(chan pageLinks)
	go func() {
		for link := range links {
			if link.Seen {
				continue
			}
			fmt.Printf("Sending %v to parse\n", link.Page)
			select {
			case <-done:
				return
			case toParse <- link:
				fmt.Printf("Sent %v to parse\n", link.Page)
			}
		}
		close(toParse)
	}()

	return toParse
}

func parsePage(done <-chan bool, links <-chan pageLinks, toLoad chan<- pageLinks, parserId int) <-chan int {
	parsed := make(chan int)
	go func() {
		for page := range links {
			if page.Seen {
				continue
			}
			page.Seen = true

			fmt.Printf("Parser %v Parsing page %v\n", parserId, page.Page)
			time.Sleep(time.Millisecond * 250)
			for _, link := range page.Links {
				if link.Seen {
					continue
				}
				fmt.Printf("Sending %v to load\n", link.Page)
				toLoad <- link
				fmt.Printf("Sent %v to load\n", link.Page)
			}

			fmt.Printf("Sending %v to parsed\n", page.Page)
			select {
			case <-done:
				return
			case parsed <- page.Page:
				fmt.Printf("Sent %v to parsed\n", page.Page)
			}
		}
		close(parsed)
	}()
	return parsed
}

func fanIn(done <-chan bool, parsedChannels ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	wg.Add(len(parsedChannels))

	parsedPages := make(chan int)
	for _, parsed := range parsedChannels {
		go func(parsed <-chan int) {
			defer wg.Done()

			for p := range parsed {

				fmt.Printf("Sending %v to parsed pages\n", p)
				select {
				case <-done:
					return
				case parsedPages <- p:
					fmt.Printf("Sent %v to parsed pages\n", p)
				}
			}
		}(parsed)
	}

	go func() {
		wg.Wait()
		close(parsedPages)
	}()
	return parsedPages
}

func (c Crawler) Crawl() {
	done := make(chan bool)
	defer close(done)

	toLoad := make(chan pageLinks)
	toParse := queuePages(done, toLoad)

	parsed := make([]<-chan int, lib.MaxParsers)
	for i := 0; i < lib.MaxParsers; i++ {
		parsed[i] = parsePage(done, toParse, toLoad, i)
	}
	toLoad <- c.RootPage

	for parsedPage := range fanIn(done, parsed...) {
		fmt.Printf("Received parsed page %v\n", parsedPage)
	}
}
