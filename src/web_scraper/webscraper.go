package web_scraper

import (
	"fmt"
	"github.com/gocolly/colly/v2"
)

// https://do512.com/
func gatherDo512(){
	c := colly.NewCollector()

	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		err := e.Request.Visit(e.Attr("href"))
		if err != nil {
			return
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	err := c.Visit("https://go-colly.org/")
	if err != nil {
		return
	}
}