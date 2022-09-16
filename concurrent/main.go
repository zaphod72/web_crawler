package main

import (
	"fmt"
	"math/rand"

	"github.com/zaphod72/web_crawler/concurrent/lib"
	"github.com/zaphod72/web_crawler/concurrent/waitgroup"
	"github.com/zaphod72/web_crawler/concurrent/worker_pool"
)

const totalPages = 100

func main() {
	pages := make([]lib.PageLinks, totalPages)

	for i := 0; i < totalPages; i++ {
		nLinks := rand.Intn(3) + 1
		pages[i].Page = i
		pages[i].Links = make([]*lib.PageLinks, nLinks)

		for n := 0; n < nLinks; n++ {
			pages[i].Links[n] = &pages[rand.Intn(totalPages)]
		}
	}

	pageCrawler := lib.PageCrawler{&pages[0]}
	workerPoolCrawler := worker_pool.Crawler{pageCrawler}
	waitgroupCrawler := waitgroup.Crawler{pageCrawler}
	workerPoolCrawler.Crawl()

	verify(pages)
}

func verify(pages []lib.PageLinks) {
	crawled := make([]bool, totalPages)
	listCrawled(&pages[0], crawled)
	for i := 0; i < totalPages; i++ {
		if crawled[i] != pages[i].Seen {
			fmt.Printf("Page %v, expected %v, actual %v\n", i, crawled[i], pages[i].Seen)
		}
	}
}

func listCrawled(rootPage *lib.PageLinks, crawled []bool) {
	crawled[rootPage.Page] = true
	queue := make([]*lib.PageLinks, 0, 20)
	queue = append(queue, rootPage)

	for len(queue) > 0 {
		curPage := queue[0]

		for _, link := range curPage.Links {
			if !crawled[link.Page] {
				crawled[link.Page] = true
				queue = append(queue, link)
			}
		}

		queue = queue[1:]
	}
}
