package scraper

import (
	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"github.com/PuerkitoBio/goquery"
	"log"
)

func (s *Scraper) extractChaptersFromPage(
	URL string, chapterSelectors []string, contentSelector string, chapterURLs *[]string,
) (appendedChapter bool) {
	res, err := s.session.Get(URL)
	raven.CheckError(err)
	doc := s.session.GetDocument(res)
	doc.Find(chapterSelectors[0]).Each(func(i int, selection *goquery.Selection) {
		chapterLink, exists := selection.Attr("href")
		if !exists {
			log.Fatal("chapter selectors have to select a link!")
		}
		// final selector, if it has a link it should be the chapter
		if len(chapterSelectors) == 1 {
			*chapterURLs = append(*chapterURLs, chapterLink)
			appendedChapter = true
			return
		}
		// check if child could append chapter, if not append the current link
		childAppendedChapter := s.extractChaptersFromPage(
			chapterLink,
			chapterSelectors[1:],
			contentSelector,
			chapterURLs,
		)
		// if direct child didn't append a chapter link and we find something with the content selector
		// we append the current link to our chapter list
		if !childAppendedChapter && doc.Find(contentSelector).Length() > 0 {
			*chapterURLs = append(*chapterURLs, chapterLink)
			appendedChapter = true
		}
	})
	return appendedChapter
}
