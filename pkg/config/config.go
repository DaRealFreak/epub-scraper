package config

import (
	"net/url"

	"github.com/DaRealFreak/epub-scraper/pkg/raven"
)

// NovelConfig contains the configuration of the novel scraper
type NovelConfig struct {
	BaseDirectory string
	General       General             `yaml:"general"`
	Sites         []SiteConfiguration `yaml:"sites"`
	Chapters      []Source            `yaml:"chapters"`
	Assets        Assets              `yaml:"assets"`
	BackList      []string            `yaml:"blacklist"`
}

// TitleContent contains the title selector and prefix/suffix selectors
// also offers a regular expression cleanup of the title
type TitleContent struct {
	AddPrefix       *bool     `yaml:"add-prefix"`
	TitleSelector   *string   `yaml:"title-selector"`
	PrefixSelectors *[]string `yaml:"prefix-selectors"`
	SuffixSelectors *[]string `yaml:"suffix-selectors"`
	CleanupRegex    string    `yaml:"cleanup-regex"`
}

// ChapterContent contains the content selector and the author note end selector
// nearly always at the start of the chapter, so start f.e. on the title
type ChapterContent struct {
	ContentSelector *string   `yaml:"content-selector"`
	PrefixSelectors *[]string `yaml:"prefix-selectors"`
	SuffixSelectors *[]string `yaml:"suffix-selectors"`
}

// Pagination contains all implemented options for paginations of websites
type Pagination struct {
	ReversePosts     *bool   `yaml:"reverse-posts"`
	NextPageSelector *string `yaml:"next-page-selector"`
}

// GetSiteConfigFromURL retrieves the site configuration for the passed URL
// will return an empty site configuration with nil values if no site configuration for host exists
func (s *NovelConfig) GetSiteConfigFromURL(url *url.URL) *SiteConfiguration {
	for _, site := range s.Sites {
		if url.Host == site.Host {
			// pin variable for scope linting
			site := site
			return &site
		}
	}
	// return empty configuration with nil values
	return &SiteConfiguration{}
}

// IsURLBlacklisted checks if the passed URL is blacklisted
// it parses the passed URL and the blacklisted URLs to ignore minor differences like f.e. trailing slash
func (s *NovelConfig) IsURLBlacklisted(checkedURL string) bool {
	check, err := url.Parse(checkedURL)
	raven.CheckError(err)
	for _, listItem := range s.BackList {
		parsedListItem, err := url.Parse(listItem)
		raven.CheckError(err)
		if check.String() == parsedListItem.String() {
			return true
		}
	}
	return false
}
