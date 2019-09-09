package scraper

import (
	"strings"

	"github.com/DaRealFreak/epub-scraper/pkg/config"
	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"github.com/PuerkitoBio/goquery"
)

func (s *Scraper) getChapterContent(chapterURL string, content config.ChapterContent) string {
	res, err := s.session.Get(chapterURL)
	raven.CheckError(err)

	doc := s.session.GetDocument(res)
	chapterContent, err := doc.Find(content.ContentSelector).First().Html()
	raven.CheckError(err)

	for _, authorNoteSelector := range content.AuthorNoteEndSelector {
		chapterContent = s.removeAuthorBlock(chapterContent, authorNoteSelector)
	}

	for _, authorNoteSelector := range content.FooterStartSelector {
		chapterContent = s.removeFooterBlock(chapterContent, authorNoteSelector)
	}

	return chapterContent
}

func (s *Scraper) removeAuthorBlock(chapterContent string, selector string) string {
	contentDoc, err := goquery.NewDocumentFromReader(strings.NewReader(chapterContent))
	raven.CheckError(err)
	contentDoc.Find(selector).Each(func(i int, selection *goquery.Selection) {
		afterAuthor, err := selection.Html()
		raven.CheckError(err)
		chapterContent = strings.Join(strings.Split(chapterContent, afterAuthor)[1:], "")
	})
	return chapterContent
}

func (s *Scraper) removeFooterBlock(chapterContent string, selector string) string {
	contentDoc, err := goquery.NewDocumentFromReader(strings.NewReader(chapterContent))
	raven.CheckError(err)
	contentDoc.Find(selector).Each(func(i int, selection *goquery.Selection) {
		afterAuthor, err := selection.Html()
		raven.CheckError(err)
		chapterContent = strings.Split(chapterContent, afterAuthor)[0]
	})
	return chapterContent
}
