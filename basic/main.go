package main

import (
	"fmt"
	"math/rand"
	"time"
)

const nPages = 100

type void struct{}

var member void

type pageLinks struct {
	Page  int
	Links []*pageLinks
}

func crawl(rootPage pageLinks) {
	linksSeen := make(map[int]void, nPages)
	queue := make([]*pageLinks, 0, 20)
	linksSeen[rootPage.Page] = member
	queue = append(queue, &rootPage)

	for len(queue) > 0 {
		curPage := queue[0]

		for _, link := range getLinks(curPage) {
			_, ok := linksSeen[link.Page]
			if !ok {
				linksSeen[link.Page] = member
				queue = append(queue, link)
			}
		}

		queue = queue[1:]
	}
}

func getLinks(page *pageLinks) []*pageLinks {
	fmt.Printf("Parsing %v\n", page.Page)
	time.Sleep(time.Millisecond * 250)
	return page.Links
}

func main() {
	pages := make([]pageLinks, nPages)

	for i := 0; i < nPages; i++ {
		nLinks := rand.Intn(15) + 5
		pages[i].Page = i
		pages[i].Links = make([]*pageLinks, nLinks)

		for n := 0; n < nLinks; n++ {
			pages[i].Links[n] = &pages[rand.Intn(nPages)]
		}
	}

	crawl(pages[0])
}
