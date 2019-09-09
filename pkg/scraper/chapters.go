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
	html, err := doc.Find(content.ContentSelector).First().Html()
	raven.CheckError(err)

	contentDoc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	raven.CheckError(err)
	contentDoc.Find(content.AuthorNoteEndSelector).Each(func(i int, selection *goquery.Selection) {
		afterAuthor, err := selection.Html()
		raven.CheckError(err)
		html = strings.Join(strings.Split(html, afterAuthor)[1:], "")
	})
	return html
}
