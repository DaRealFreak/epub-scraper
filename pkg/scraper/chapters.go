package scraper

import (
	"regexp"
	"strings"

	"github.com/DaRealFreak/epub-scraper/pkg/config"
	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

// extractChapterData extracts the Chapter data from the passed final chapter URL (no more redirects, only content)
func (s *Scraper) extractChapterData(
	chapterURL string, titleConfig config.TitleContent, chapterConfig config.ChapterContent,
) *ChapterData {
	log.Infof("extracting title and content from URL: %s", chapterURL)
	res, err := s.session.Get(chapterURL)
	raven.CheckError(err)
	doc := s.session.GetDocument(res)

	return &ChapterData{
		addPrefix: *titleConfig.AddPrefix,
		title:     s.getChapterTitle(doc, &titleConfig),
		content:   s.getChapterContent(doc, &chapterConfig),
	}
}

// getChapterContent returns the chapter content of the passed URL based on the passed ChapterContent settings
func (s *Scraper) getChapterContent(doc *goquery.Document, content *config.ChapterContent) string {
	chapterContent, err := doc.Find(*content.ContentSelector).First().Html()
	raven.CheckError(err)

	for _, prefixSelector := range *content.PrefixSelectors {
		chapterContent = s.removePrefix(chapterContent, prefixSelector)
	}

	for _, suffixSelector := range *content.SuffixSelectors {
		chapterContent = s.removeSuffix(chapterContent, suffixSelector)
	}

	chapterContent = s.fixHTMLCode(chapterContent)
	chapterContent = s.sanitizer.Sanitize(chapterContent)
	return chapterContent
}

// removePrefix removes the author block of the extracted chapter content based on the selector
func (s *Scraper) removePrefix(chapterContent string, selector string) string {
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

// removeSuffix removes the footer block of the extracted chapter content based on the selector
func (s *Scraper) removeSuffix(chapterContent string, selector string) string {
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

// getChapterTitle returns the chapter title of the passed URL based on the passed ChapterContent settings
func (s *Scraper) getChapterTitle(doc *goquery.Document, content *config.TitleContent) string {
	titleContent, err := doc.Html()
	raven.CheckError(err)

	// ToDo: use document.Find(sel).First().NextAll() instead of ripping apart the HTML
	for _, prefixSelector := range *content.PrefixSelectors {
		titleContent = s.removePrefix(titleContent, prefixSelector)
	}

	for _, suffixSelector := range *content.SuffixSelectors {
		titleContent = s.removeSuffix(titleContent, suffixSelector)
	}

	doc, err = goquery.NewDocumentFromReader(strings.NewReader(titleContent))
	raven.CheckError(err)

	title := doc.Find(*content.TitleSelector).First().Text()

	// cleanup title if cleanup regular expression is given in the configuration
	if content.CleanupRegex != "" {
		re := regexp.MustCompile(content.CleanupRegex)
		matches := re.FindStringSubmatch(title)

		paramsMap := make(map[string]string)
		for i, name := range re.SubexpNames() {
			if i > 0 && i <= len(matches) {
				paramsMap[name] = matches[i]
			}
		}

		if val, ok := paramsMap["Title"]; ok {
			title = val
		} else {
			log.Fatal("capture group \"Title\" is required for the title cleanup pattern")
		}
	}
	return title
}
