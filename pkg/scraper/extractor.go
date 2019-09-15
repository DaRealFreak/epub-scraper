package scraper

import (
	"net/url"

	"github.com/DaRealFreak/epub-scraper/pkg/config"
	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

// tocContent contains all relevant information for extracting chapter data
type tocContent struct {
	toc      *config.Toc
	cfg      *config.NovelConfig
	chapters []*ChapterData
}

// handleToc handles passed Table of Content configurations to extract the Chapter data
func (s *Scraper) handleToc(toc *config.Toc, cfg *config.NovelConfig) (chapters []*ChapterData) {
	content := &tocContent{
		toc:      toc,
		cfg:      cfg,
		chapters: []*ChapterData{},
	}
	s.navigateThroughToc(toc.URL, content)

	if *toc.Pagination.ReversePosts {
		for i, j := 0, len(content.chapters)-1; i < j; i, j = i+1, j-1 {
			content.chapters[i], content.chapters[j] = content.chapters[j], content.chapters[i]
		}
	}
	return content.chapters
}

// navigateThroughToc navigates through the table of content and extracts chapter links
func (s *Scraper) navigateThroughToc(tocURL string, content *tocContent) {
	base, err := url.Parse(tocURL)
	raven.CheckError(err)
	res, err := s.session.Get(base.String())
	raven.CheckError(err)
	doc := s.session.GetDocument(res)
	// extract and append chapter URLs from the current page
	log.Infof("extracting chapters from %s", tocURL)
	s.extractChapters(base, doc, content)

	// if we have a pagination check for next page and repeat the process
	if content.toc.Pagination.NextPageSelector != nil {
		doc.Find(*content.toc.Pagination.NextPageSelector).Each(func(i int, selection *goquery.Selection) {
			tocPage, exists := selection.Attr("href")
			if !exists {
				log.Fatal("chapter selectors have to select a link!")
			}
			// resolve reference to parse relative strings
			u, err := url.Parse(tocPage)
			raven.CheckError(err)
			tocPage = base.ResolveReference(u).String()

			// prevent infinite loop to same page
			if tocURL != tocPage {
				s.navigateThroughToc(tocPage, content)
			}
		})
	}
}

// extractChapters extracts possible chapters from the ToC page and resolves the redirects
func (s *Scraper) extractChapters(base *url.URL, doc *goquery.Document, content *tocContent) {
	doc.Find(content.toc.ChapterSelector).Each(func(i int, selection *goquery.Selection) {
		chapterURL, exists := selection.Attr("href")
		if !exists {
			log.Warningf("no chapter URL found in: %s", base.String())
			return
		}
		u, err := url.Parse(chapterURL)
		raven.CheckError(err)
		chapterURL = base.ResolveReference(u).String()
		chapterData := s.extractChapterData(chapterURL, content.cfg, content.toc.SourceContent)
		content.chapters = append(content.chapters, chapterData)
	})
}
