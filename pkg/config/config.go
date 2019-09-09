package config

// General contains the general information about the novel
type General struct {
	Title       string        `yaml:"title"`
	AltTitle    string        `yaml:"alt-title"`
	Author      string        `yaml:"author"`
	Cover       string        `yaml:"cover"`
	Raw         string        `yaml:"raw"`
	Translators []*Translator `yaml:"translators"`
}

// Translator contains the name and website of the translators
type Translator struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

// URL contains the chapter specific options like the chapter url, selector and content options
type URL struct {
	URL             string         `yaml:"url"`
	ChapterSelector string         `yaml:"chapter-selector"`
	ChapterContent  ChapterContent `yaml:"chapter-content"`
}

// ChapterContent contains the content selector and the author note end selector
// nearly always at the start of the chapter, so start f.e. on the title
type ChapterContent struct {
	ContentSelector       string   `yaml:"content-selector"`
	AuthorNoteEndSelector []string `yaml:"author-note-end-selector"`
	FooterStartSelector   []string `yaml:"footer-start-selector"`
}

// Toc contains all relevant information to extract the content of the novel chapters
type Toc struct {
	URLs []*URL `yaml:"urls"`
}

// Assets contains the included assets in each added section
type Assets struct {
	CSS  Asset `yaml:"css"`
	Font Asset `yaml:"fonts"`
}

// Asset contains the path on the host system and after being added the internal path of the epub
type Asset struct {
	HostPath     string `yaml:"path"`
	InternalPath string
}

// NovelConfig contains the configuration of the novel scraper
type NovelConfig struct {
	BaseDirectory string
	General       General `yaml:"general"`
	Toc           Toc     `yaml:"toc"`
	Assets        Assets  `yaml:"assets"`
}
