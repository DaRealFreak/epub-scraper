package config

// NovelConfig contains the configuration of the novel scraper
type NovelConfig struct {
	BaseDirectory string
	General       General             `yaml:"general"`
	Sites         []SiteConfiguration `yaml:"sites"`
	Chapters      []Source            `yaml:"chapters"`
	Assets        Assets              `yaml:"assets"`
}

// ChapterContent contains the content selector and the author note end selector
// nearly always at the start of the chapter, so start f.e. on the title
type ChapterContent struct {
	AddPrefix             bool     `yaml:"add-prefix"`
	TitleSelector         string   `yaml:"title-selector"`
	ContentSelector       string   `yaml:"content-selector"`
	AuthorNoteEndSelector []string `yaml:"author-note-end-selector"`
	FooterStartSelector   []string `yaml:"footer-start-selector"`
}

// Pagination contains all implemented options for paginations of websites
type Pagination struct {
	ReversePosts     bool     `yaml:"reverse-posts"`
	NextPageSelector string   `yaml:"next-page-selector"`
}
