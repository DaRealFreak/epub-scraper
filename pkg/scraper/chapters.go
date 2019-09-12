package scraper

import (
	"strings"

	"github.com/DaRealFreak/epub-scraper/pkg/config"
	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"github.com/PuerkitoBio/goquery"
)

// getChapterContent returns the chapter content of the passed URL based on the passed ChapterContent settings
func (s *Scraper) getChapterContent(chapterURL string, content *config.ChapterContent) string {
	res, err := s.session.Get(chapterURL)
	raven.CheckError(err)

	doc := s.session.GetDocument(res)
	chapterContent, err := doc.Find(*content.ContentSelector).First().Html()
	raven.CheckError(err)

	for _, authorNoteSelector := range *content.AuthorNoteEndSelector {
		chapterContent = s.removeAuthorBlock(chapterContent, authorNoteSelector)
	}

	for _, authorNoteSelector := range *content.FooterStartSelector {
		chapterContent = s.removeFooterBlock(chapterContent, authorNoteSelector)
	}

	chapterContent = s.fixHTMLCode(chapterContent)
	chapterContent = s.sanitizer.Sanitize(chapterContent)
	return chapterContent
}

// removeAuthorBlock removes the author block of the extracted chapter content based on the selector
func (s *Scraper) removeAuthorBlock(chapterContent string, selector string) string {
	contentDoc, err := goquery.NewDocumentFromReader(strings.NewReader(chapterContent))
	raven.CheckError(err)
	selection := contentDoc.Find(selector).First()
	if selection.Length() > 0 {
		afterAuthor, err := goquery.OuterHtml(selection)
		raven.CheckError(err)
		chapterContent = strings.Join(strings.Split(chapterContent, afterAuthor)[1:], "")
	}
	return chapterContent
}

// removeFooterBlock removes the footer block of the extracted chapter content based on the selector
func (s *Scraper) removeFooterBlock(chapterContent string, selector string) string {
	contentDoc, err := goquery.NewDocumentFromReader(strings.NewReader(chapterContent))
	raven.CheckError(err)
	selection := contentDoc.Find(selector).First()
	if selection.Length() > 0 {
		afterFooter, err := goquery.OuterHtml(selection)
		raven.CheckError(err)
		chapterContent = strings.Split(chapterContent, afterFooter)[0]
	}
	return chapterContent
}
