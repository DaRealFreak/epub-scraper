package scraper

import (
	"fmt"
	"net/url"

	"github.com/DaRealFreak/epub-scraper/pkg/config"
	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

// tocContent contains all relevant information for extracting chapter data
type tocContent struct {
	toc         *config.Toc
	cfg         *config.NovelConfig
	chapterUrls []string
}

// handleToc handles passed Table of Content configurations to extract the Chapter data
func (s *Scraper) handleToc(toc *config.Toc, cfg *config.NovelConfig) (chapters []*ChapterData) {
	content := &tocContent{
		toc:         toc,
		cfg:         cfg,
		chapterUrls: []string{},
	}
	s.navigateThroughToc(toc.URL, content)

	if *toc.Pagination.ReversePosts {
		for i, j := 0, len(content.chapterUrls)-1; i < j; i, j = i+1, j-1 {
			content.chapterUrls[i], content.chapterUrls[j] = content.chapterUrls[j], content.chapterUrls[i]
		}
	}

	fmt.Println(content.chapterUrls)
	for _, chapterURL := range content.chapterUrls {
		chapters = append(chapters, s.extractChapterData(chapterURL, toc.TitleContent, toc.ChapterContent))
	}
	return chapters
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
		chapterURL = s.resolveRedirects(chapterURL, content)
		log.Infof("adding chapter URL to list: %s", chapterURL)
		content.chapterUrls = append(content.chapterUrls, chapterURL)
	})
}

// resolveRedirects calls the passed URL to check for any 30x status codes (redirect)
// and resolves the redirect instructions from the configuration
func (s *Scraper) resolveRedirects(chapterURL string, content *tocContent) (resolvedChapterURL string) {
	res, err := s.session.Get(chapterURL)
	raven.CheckError(err)
	// follow redirects for f.e. exit links from novelupdates
	chapterURL = res.Request.URL.String()
	// retrieve site config for the host of the chapter url
	siteConfig := content.cfg.GetSiteConfigFromURL(res.Request.URL)
	// if we have redirects resolve them
	if len(siteConfig.Redirects) > 0 {
		doc := s.session.GetDocument(res)
		for _, redirect := range siteConfig.Redirects {
			log.Debugf("resolving redirects for URL: %s", chapterURL)
			// iterate through every redirect and update the link
			redirectLink, exists := doc.Find(redirect).First().Attr("href")
			// no redirect found use the URL from before
			if !exists {
				log.Debugf("could not find redirect for selector: %s", redirect)
				break
			}
			// request the found redirect link and update the document
			res, err := s.session.Get(redirectLink)
			raven.CheckError(err)
			doc = s.session.GetDocument(res)
			// update chapter URL from the request URL in case we get redirected
			chapterURL = res.Request.URL.String()
			log.Debugf("got redirected to url: %s", chapterURL)
			// break in case we got redirected to a different host
			if res.Request.Host != siteConfig.Host {
				break
			}
		}
		parsedURL, err := url.Parse(chapterURL)
		raven.CheckError(err)
		// if the redirected URL has a different host resolve the redirects from the new host too
		if parsedURL.Host != siteConfig.Host {
			return s.resolveRedirects(chapterURL, content)
		}
	}
	return chapterURL
}
