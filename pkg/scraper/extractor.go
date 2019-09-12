package scraper

import (
	"github.com/DaRealFreak/epub-scraper/pkg/config"
	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

type ChapterData struct {
	addPrefix bool
	title     string
	content   string
}

// handleToc handles passed Table of Content configurations to extract the Chapter data
func (s *Scraper) handleToc(toc *config.Toc) (chapters []*ChapterData) {
	// extract all chapters recursively from the passed ToC URL
	var chapterUrls []string
	s.extractTocPages(toc.URL, toc, &chapterUrls)

	if *toc.Pagination.ReversePosts {
		for i, j := 0, len(chapterUrls)-1; i < j; i, j = i+1, j-1 {
			chapterUrls[i], chapterUrls[j] = chapterUrls[j], chapterUrls[i]
		}
	}

	for _, chapterURL := range chapterUrls {
		chapters = append(chapters, s.extractChapterData(chapterURL, toc.TitleContent, toc.ChapterContent))
	}
	return chapters
}

// extractTocPages extracts further ToC pages based on the pagination settings
// it'll automatically skip if the redirected URL equals the current URL
func (s *Scraper) extractTocPages(url string, toc *config.Toc, chapterUrls *[]string) {
	res, err := s.session.Get(url)
	raven.CheckError(err)
	doc := s.session.GetDocument(res)
	// extract and append chapter URLs from the current page
	s.extractChaptersFromPage(doc, toc.ChapterSelectors, *toc.ChapterContent.ContentSelector, chapterUrls)

	// if we have a pagination check for next page and repeat the process
	if toc.Pagination.NextPageSelector != nil {
		doc.Find(*toc.Pagination.NextPageSelector).Each(func(i int, selection *goquery.Selection) {
			tocPage, exists := selection.Attr("href")
			if !exists {
				log.Fatal("chapter selectors have to select a link!")
			}
			// prevent infinite loop to same page
			if url != tocPage {
				s.extractTocPages(tocPage, toc, chapterUrls)
			}
		})
	}

}

// extractChaptersFromPage extracts all available chapters from the passed ToC URl
// and recursively follows possible redirects
// if the last level does not return any matches it'll add the closest level where the chapter content can be found
func (s *Scraper) extractChaptersFromPage(
	doc *goquery.Document, chapterSelectors []string, contentSelector string, chapterURLs *[]string,
) (appendedChapter bool) {
	doc.Find(chapterSelectors[0]).Each(func(i int, selection *goquery.Selection) {
		// to follow a link or add a chapter link a link has to be selected, else log this as fatal and exit
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

		res, err := s.session.Get(chapterLink)
		raven.CheckError(err)
		// check if child could append chapter
		childAppendedChapter := s.extractChaptersFromPage(
			s.session.GetDocument(res),
			chapterSelectors[1:],
			contentSelector,
			chapterURLs,
		)
		// if children didn't append a chapter link and we find matches with the content selector
		// we append the current link to our chapter list and return true to the parents
		if !childAppendedChapter && doc.Find(contentSelector).Length() > 0 {
			*chapterURLs = append(*chapterURLs, chapterLink)
			appendedChapter = true
		}
	})
	return appendedChapter
}
