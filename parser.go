package main

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func getLinks(body string, selector string) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	var links []string
	doc.Find(selector).Each(func(index int, item *goquery.Selection) {
		link, _ := item.Attr("href")
		link = strings.Trim(link, "\r\n ")
		if len(link) > 0 {
			links = append(links, link)
		}
	})
	return links, nil
}

func getPages(body string) ([]string, error) {
	return getLinks(body, ".page_numbers a")
}

func getPageLinks(body string) ([]string, error) {
	return getLinks(body, ".imd_photo a")
}
