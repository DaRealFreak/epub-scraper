package scraper

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/DaRealFreak/epub-scraper/pkg/config"
	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"github.com/DaRealFreak/epub-scraper/pkg/unicode"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

// extractChapterData follows the redirects from the URL and the configuration
// and extracts the chapter title/content from the final URL
func (s *Scraper) extractChapterData(
	chapterURL string, cfg *config.NovelConfig, srcCfg config.SourceContent,
) (chapterData *ChapterData) {
	// directly return nil if initial URL is blacklisted
	if cfg.IsURLBlacklisted(chapterURL) {
		return nil
	}

	res, err := s.session.Get(chapterURL)
	raven.CheckError(err)
	doc := s.session.GetDocument(res)
	// follow redirects for f.e. exit links from novelupdates
	// retrieve site config for the host of the chapter url
	siteConfig := cfg.GetSiteConfigFromURL(res.Request.URL)

	if !s.isURLEqual(chapterURL, res.Request.URL.String()) {
		// update configuration to match the new host
		chapterURL = res.Request.URL.String()
		srcCfg = siteConfig.SourceContent
		log.Debugf("got redirected to url: %s", chapterURL)
	}

	// if we have redirects resolve them
	if len(siteConfig.Redirects) > 0 {
		for _, redirect := range siteConfig.Redirects {
			log.Debugf("resolving redirects for URL: %s", chapterURL)
			// iterate through every redirect and update the link
			redirectLink, exists := doc.Find(redirect).First().Attr("href")
			// handle eventual relative URLs
			redirectURL, err := url.Parse(redirectLink)
			raven.CheckError(err)
			redirectLink = res.Request.URL.ResolveReference(redirectURL).String()
			// no redirect found use the URL from before
			if !exists {
				log.Debugf("could not find any redirect for selector: %s", redirect)
				break
			}
			if cfg.IsURLBlacklisted(redirectLink) {
				break
			}
			// request the found redirect link and update the document we will use for the chapter extraction
			res, err = s.session.Get(redirectLink)
			raven.CheckError(err)
			doc = s.session.GetDocument(res)
			log.Debugf("got redirected to url: %s", res.Request.URL.String())
			// break in case we got redirected to a different host
			if res.Request.Host != siteConfig.Host {
				srcCfg = cfg.GetSiteConfigFromURL(res.Request.URL).SourceContent
				break
			}
		}
		// if the redirected URL has a different host resolve the redirects from the new host too
		parsedURL, err := url.Parse(chapterURL)
		raven.CheckError(err)
		if parsedURL.Host != siteConfig.Host {
			// update configuration to match the new host
			srcCfg = siteConfig.SourceContent
			return s.extractChapterData(chapterURL, cfg, srcCfg)
		}
	}
	finalChapterURL := res.Request.URL.String()
	if cfg.IsURLBlacklisted(finalChapterURL) {
		return nil
	}
	log.Infof("extracting chapter from %s", finalChapterURL)
	chapterData = &ChapterData{
		addPrefix: *srcCfg.TitleContent.AddPrefix,
		title:     s.getChapterTitle(doc, &srcCfg.TitleContent),
		content:   s.getChapterContent(doc, &srcCfg.ChapterContent),
	}
	log.Infof("extracted chapter: %s (content length: %d)", chapterData.title, len(chapterData.content))
	return chapterData
}

// getChapterContent returns the chapter content of the passed URL based on the passed ChapterContent settings
func (s *Scraper) getChapterContent(doc *goquery.Document, content *config.ChapterContent) string {
	chapterContent, err := doc.Find(*content.ContentSelector).First().Html()
	raven.CheckError(err)

	chapterContent = s.applyCleanupOptions(chapterContent, &content.CleanupOptions, "Content")
	return unicode.StripUnicodeEmojis(s.sanitizer.Sanitize(s.fixHTMLCode(chapterContent)))
}

// getChapterTitle returns the chapter title of the passed URL based on the passed ChapterContent settings
func (s *Scraper) getChapterTitle(doc *goquery.Document, content *config.TitleContent) string {
	// if we only use the prefix the title selector can be nil too
	if content.TitleSelector == nil {
		return ""
	}

	titleContent, err := doc.Find(*content.TitleSelector).First().Html()
	raven.CheckError(err)

	titleContent = s.applyCleanupOptions(titleContent, &content.CleanupOptions, "Title")

	doc, err = goquery.NewDocumentFromReader(strings.NewReader(titleContent))
	raven.CheckError(err)
	return strings.TrimSpace(unicode.StripUnicodeEmojis(doc.Text()))
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

// applyCleanupOptions applies the cleanup options to the passed html before returning it again
func (s *Scraper) applyCleanupOptions(htmlContent string, options *config.CleanupOptions, captureGroup string) string {
	// strip unicode emojis from the title and trim the text before parsing with the regular expressions
	htmlContent = unicode.SanitizeSpaces(htmlContent)

	// ToDo: use document.Find(sel).First().NextAll() instead of ripping apart the HTML
	if options.PrefixSelectors != nil {
		for _, prefixSelector := range *options.PrefixSelectors {
			htmlContent = s.removePrefix(htmlContent, prefixSelector)
		}
	}

	if options.SuffixSelectors != nil {
		for _, suffixSelector := range *options.SuffixSelectors {
			htmlContent = s.removeSuffix(htmlContent, suffixSelector)
		}
	}

	// strip title with regular expressions if set in the related configuration
	// (for f.e. additional notes or we have to select title from main content)
	if options.StripRegex != "" {
		re := regexp.MustCompile(options.StripRegex)
		matches := re.FindStringSubmatch(htmlContent)

		paramsMap := make(map[string]string)
		for i, name := range re.SubexpNames() {
			if i > 0 && i <= len(matches) {
				paramsMap[name] = matches[i]
			}
		}

		if val, ok := paramsMap[captureGroup]; ok {
			htmlContent = val
		} else {
			log.Fatalf("capture group %s is required for the title cleanup pattern", captureGroup)
		}
	}

	// clean up content with regular expressions if set in the related configuration (for f.e. translator notes)
	if options.CleanupRegex != "" {
		re := regexp.MustCompile(options.CleanupRegex)
		htmlContent = re.ReplaceAllString(htmlContent, "")
	}

	return htmlContent
}

// isURLEqual compares the passed URLs for equality ignoring scheme differences
func (s *Scraper) isURLEqual(url1 string, url2 string) bool {
	parsedURL1, err1 := url.Parse(url1)
	parsedURL2, err2 := url.Parse(url2)
	if err1 != nil || err2 != nil {
		return false
	}
	// set both URL schemes to https to ignore https redirects
	parsedURL1.Scheme = "https"
	parsedURL2.Scheme = "https"
	return parsedURL1.String() == parsedURL2.String()
}
