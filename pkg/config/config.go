package config

import (
	"net/url"

	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	log "github.com/sirupsen/logrus"
)

// NovelConfig contains the configuration of the novel scraper
type NovelConfig struct {
	BaseDirectory string
	General       General             `yaml:"general"`
	Sites         []SiteConfiguration `yaml:"sites"`
	Chapters      []Source            `yaml:"chapters"`
	Assets        Assets              `yaml:"assets"`
	BackList      []string            `yaml:"blacklist"`
	Replacements  []Replacement       `yaml:"replacements"`
	Templates     Templates           `yaml:"templates"`
}

// TitleContent contains the title selector and the title cleanup options
type TitleContent struct {
	AddPrefix      *bool   `yaml:"add-prefix"`
	TitleSelector  *string `yaml:"title-selector"`
	CleanupOptions `yaml:",inline"`
}

// ChapterContent contains the content selector the content cleanup options
type ChapterContent struct {
	ContentSelector *string `yaml:"content-selector"`
	CleanupOptions  `yaml:",inline"`
}

// CleanupOptions are all options related to cleaning up the extracted content of titles and chapters
type CleanupOptions struct {
	PrefixSelectors *[]string `yaml:"prefix-selectors"`
	SuffixSelectors *[]string `yaml:"suffix-selectors"`
	StripRegex      string    `yaml:"strip-regex"`
	CleanupRegex    string    `yaml:"cleanup-regex"`
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
			log.Infof("url %s is blacklisted, skipping", check.String())
			return true
		}
	}
	return false
}

// DoURLReplacements checks if the passed URL is getting replaced through the configuration
func (s *NovelConfig) DoURLReplacements(checkedURL string) (chapterUrl string, changed bool) {
	check, err := url.Parse(checkedURL)
	raven.CheckError(err)
	for _, replacement := range s.Replacements {
		parsedReplacementURL, err := url.Parse(replacement.Url)
		raven.CheckError(err)
		if check.String() == parsedReplacementURL.String() {
			log.Infof("url %s is getting replaced to %s", check.String(), replacement.ReplacementURL)
			return replacement.ReplacementURL, true
		}
	}
	return check.String(), false
}
