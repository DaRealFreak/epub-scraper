package scraper

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/DaRealFreak/epub-scraper/pkg/config"
	"github.com/DaRealFreak/epub-scraper/pkg/emojis"
	"github.com/DaRealFreak/epub-scraper/pkg/raven"
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

	if chapterURL != res.Request.URL.String() {
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
			if !exists || cfg.IsURLBlacklisted(redirectLink) {
				log.Debugf("URL is blacklisted or could not find any redirect for selector: %s", redirect)
				break
			}
			// request the found redirect link and update the document we will use for the chapter extraction
			res, err = s.session.Get(redirectLink)
			raven.CheckError(err)
			doc = s.session.GetDocument(res)
			log.Debugf("got redirected to url: %s", res.Request.URL.String())
			// break in case we got redirected to a different host
			if res.Request.Host != siteConfig.Host {
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
		log.Infof("url %s is blacklisted, skipping", finalChapterURL)
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

	if content.PrefixSelectors != nil {
		for _, prefixSelector := range *content.PrefixSelectors {
			chapterContent = s.removePrefix(chapterContent, prefixSelector)
		}
	}

	if content.SuffixSelectors != nil {
		for _, suffixSelector := range *content.SuffixSelectors {
			chapterContent = s.removeSuffix(chapterContent, suffixSelector)
		}
	}

	chapterContent = s.fixHTMLCode(chapterContent)
	chapterContent = s.sanitizer.Sanitize(chapterContent)
	chapterContent = emojis.StripUnicodeEmojis(chapterContent)
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
	if content.PrefixSelectors != nil {
		for _, prefixSelector := range *content.PrefixSelectors {
			titleContent = s.removePrefix(titleContent, prefixSelector)
		}
	}

	if content.SuffixSelectors != nil {
		for _, suffixSelector := range *content.SuffixSelectors {
			titleContent = s.removeSuffix(titleContent, suffixSelector)
		}
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
	return emojis.StripUnicodeEmojis(strings.TrimSpace(title))
}
