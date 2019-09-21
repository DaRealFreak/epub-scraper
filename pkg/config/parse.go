package config

import (
	"io/ioutil"
	"net/url"
	"path/filepath"

	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"gopkg.in/yaml.v2"
)

// Parser is a struct solely to prevent expose functions without setting up first
type Parser struct{}

// NewParser returns a pointer to an initialized parser struct
func NewParser() *Parser {
	return &Parser{}
}

// ReadConfigurationFile tries to read the passed configuration file and parse it into a NovelConfig struct
func (p *Parser) ReadConfigurationFile(fileName string) (novelConfig *NovelConfig, err error) {
	content, err := ioutil.ReadFile(filepath.Clean(fileName))
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(content, &novelConfig)
	if err != nil {
		return nil, err
	}
	// set base directory for includes and the like
	baseDirectory := filepath.Dir(fileName)
	novelConfig.BaseDirectory, err = filepath.Abs(baseDirectory)
	p.mergeSourceConfigSiteConfig(novelConfig)
	return novelConfig, err
}

// mergeSourceConfigSiteConfig merges the chapter configuration with the site configuration
// or sets the default values in case neither the chapter nor the site configuration has a value set
func (p *Parser) mergeSourceConfigSiteConfig(novelConfig *NovelConfig) {
	for _, source := range novelConfig.Chapters {
		if source.Toc != nil {
			tocURL, err := url.Parse(source.Toc.URL)
			raven.CheckError(err)
			site := novelConfig.GetSiteConfigFromURL(tocURL)
			p.updatePagination(&source.Toc.Pagination, &site.Pagination)
			p.updateTitleContent(&source.Toc.TitleContent, &site.TitleContent)
			p.updateChapterContent(&source.Toc.ChapterContent, &site.ChapterContent)
		}
		if source.Chapter != nil {
			tocURL, err := url.Parse(source.Chapter.URL)
			raven.CheckError(err)
			site := novelConfig.GetSiteConfigFromURL(tocURL)
			p.updateTitleContent(&source.Chapter.TitleContent, &site.TitleContent)
			p.updateChapterContent(&source.Chapter.ChapterContent, &site.ChapterContent)
		}
	}
}

// updatePagination updates specifically the Pagination struct of the chapter/site configuration
func (p *Parser) updatePagination(sourceConfig *Pagination, siteConfig *Pagination) {
	if sourceConfig.ReversePosts == nil {
		if siteConfig.ReversePosts == nil {
			var siteConfigDefault bool
			siteConfig.ReversePosts = &siteConfigDefault
		}
		sourceConfig.ReversePosts = siteConfig.ReversePosts
	}
	if sourceConfig.NextPageSelector == nil {
		if siteConfig.NextPageSelector == nil {
			var siteConfigDefault string
			siteConfig.NextPageSelector = &siteConfigDefault
		}
		sourceConfig.NextPageSelector = siteConfig.NextPageSelector
	}
}

// updateChapterContent updates specifically the ChapterContent struct of the chapter/site configuration
func (p *Parser) updateChapterContent(sourceConfig *ChapterContent, siteConfig *ChapterContent) {
	if sourceConfig.ContentSelector == nil {
		if siteConfig.ContentSelector == nil {
			var siteConfigDefault string
			siteConfig.ContentSelector = &siteConfigDefault
		}
		sourceConfig.ContentSelector = siteConfig.ContentSelector
	}
	if sourceConfig.CleanupOptions.PrefixSelectors == nil {
		if siteConfig.CleanupOptions.PrefixSelectors == nil {
			var siteConfigDefault []string
			siteConfig.CleanupOptions.PrefixSelectors = &siteConfigDefault
		}
		sourceConfig.CleanupOptions.PrefixSelectors = siteConfig.CleanupOptions.PrefixSelectors
	}
	if sourceConfig.CleanupOptions.SuffixSelectors == nil {
		if siteConfig.CleanupOptions.SuffixSelectors == nil {
			var siteConfigDefault []string
			siteConfig.CleanupOptions.SuffixSelectors = &siteConfigDefault
		}
		sourceConfig.CleanupOptions.SuffixSelectors = siteConfig.CleanupOptions.SuffixSelectors
	}
	if sourceConfig.CleanupOptions.StripRegex == "" && siteConfig.CleanupOptions.StripRegex != "" {
		sourceConfig.CleanupOptions.StripRegex = siteConfig.CleanupOptions.StripRegex
	}
	if sourceConfig.CleanupOptions.CleanupRegex == "" && siteConfig.CleanupOptions.CleanupRegex != "" {
		sourceConfig.CleanupOptions.CleanupRegex = siteConfig.CleanupOptions.CleanupRegex
	}
}

func (p *Parser) updateTitleContent(sourceConfig *TitleContent, siteConfig *TitleContent) {
	if sourceConfig.AddPrefix == nil {
		if siteConfig.AddPrefix == nil {
			var siteConfigDefault bool
			siteConfig.AddPrefix = &siteConfigDefault
		}
		sourceConfig.AddPrefix = siteConfig.AddPrefix
	}
	if sourceConfig.TitleSelector == nil {
		if siteConfig.TitleSelector == nil {
			var siteConfigDefault string
			siteConfig.TitleSelector = &siteConfigDefault
		}
		sourceConfig.TitleSelector = siteConfig.TitleSelector
	}
	if sourceConfig.CleanupOptions.PrefixSelectors == nil {
		if siteConfig.CleanupOptions.PrefixSelectors == nil {
			var siteConfigDefault []string
			siteConfig.CleanupOptions.PrefixSelectors = &siteConfigDefault
		}
		sourceConfig.CleanupOptions.PrefixSelectors = siteConfig.CleanupOptions.PrefixSelectors
	}
	if sourceConfig.CleanupOptions.SuffixSelectors == nil {
		if siteConfig.CleanupOptions.SuffixSelectors == nil {
			var siteConfigDefault []string
			siteConfig.CleanupOptions.SuffixSelectors = &siteConfigDefault
		}
		sourceConfig.CleanupOptions.SuffixSelectors = siteConfig.CleanupOptions.SuffixSelectors
	}
	if sourceConfig.CleanupOptions.StripRegex == "" && siteConfig.CleanupOptions.StripRegex != "" {
		sourceConfig.CleanupOptions.StripRegex = siteConfig.CleanupOptions.StripRegex
	}
	if sourceConfig.CleanupOptions.CleanupRegex == "" && siteConfig.CleanupOptions.CleanupRegex != "" {
		sourceConfig.CleanupOptions.CleanupRegex = siteConfig.CleanupOptions.CleanupRegex
	}
}
