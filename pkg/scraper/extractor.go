package scraper

import (
	"github.com/DaRealFreak/epub-scraper/pkg/config"
	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

type ChapterData struct {
	title   string
	content string
}

// handleToc handles passed Table of Content configurations to extract the Chapter data
func (s *Scraper) handleToc(toc *config.Toc) (chapters []*ChapterData) {
	// extract ToC pages from the pagination configuration
	// the ToC URL is always set as the first page
	tocPages := []string{toc.URL}
	if toc.Pagination.NextPageSelector != nil {
		s.extractTocPages(toc.URL, toc.Pagination, &tocPages)
	}

	// extract all chapters from the extracted ToC pages
	var chapterUrls []string
	for _, tocPage := range tocPages {
		s.extractChaptersFromPage(tocPage, toc.ChapterSelectors, *toc.ChapterContent.ContentSelector, &chapterUrls)
	}

	if *toc.Pagination.ReversePosts {
		for i, j := 0, len(chapterUrls)-1; i < j; i, j = i+1, j-1 {
			chapterUrls[i], chapterUrls[j] = chapterUrls[j], chapterUrls[i]
		}
	}

	for _, chapterURL := range chapterUrls {
		chapters = append(chapters, s.extractChapterContent(chapterURL, toc.ChapterContent))
	}
	return chapters
}

// handleChapter handles a single chapter configurations to extract the Chapter data
func (s *Scraper) handleChapter(chapter *config.Chapter) (chapterData *ChapterData) {
	return s.extractChapterContent(chapter.URL, chapter.ChapterContent)
}

// extractChapterContent extracts the Chapter data from the passed final chapter URL (no more redirects, only content)
func (s *Scraper) extractChapterContent(chapterURL string, config config.ChapterContent) *ChapterData {
	log.Infof("extracting content from URL: %s", chapterURL)
	return &ChapterData{
		title:   chapterURL,
		content: s.getChapterContent(chapterURL, &config),
	}
}

// extractTocPages extracts further ToC pages based on the pagination settings
// it'll automatically skip if the redirected URL equals the current URL
func (s *Scraper) extractTocPages(url string, pagination config.Pagination, tocPages *[]string) {
	res, err := s.session.Get(url)
	raven.CheckError(err)
	doc := s.session.GetDocument(res)
	doc.Find(*pagination.NextPageSelector).Each(func(i int, selection *goquery.Selection) {
		tocPage, exists := selection.Attr("href")
		if !exists {
			log.Fatal("chapter selectors have to select a link!")
		}
		// prevent infinite loop to same page
		if url != tocPage {
			*tocPages = append(*tocPages, tocPage)
			s.extractTocPages(tocPage, pagination, tocPages)
		}
	})
}

// extractChaptersFromPage extracts all available chapters from the passed ToC URl
// and recursively follows possible redirects
// if the last level does not return any matches it'll add the closest level where the chapter content can be found
func (s *Scraper) extractChaptersFromPage(
	url string, chapterSelectors []string, contentSelector string, chapterURLs *[]string,
) (appendedChapter bool) {
	res, err := s.session.Get(url)
	raven.CheckError(err)
	doc := s.session.GetDocument(res)
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
		// check if child could append chapter
		childAppendedChapter := s.extractChaptersFromPage(
			chapterLink,
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
