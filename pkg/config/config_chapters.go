package config

// Source is the option to define the source of the chapter content, table of content or single chapter
type Source struct {
	Toc     *Toc     `yaml:"toc"`
	Chapter *Chapter `yaml:"chapter"`
}

// Toc is the struct for a table of content, requires the URL and the ChapterSelectors
// and implements the Pagination struct and the ChapterContent struct
type Toc struct {
	URL              string   `yaml:"url"`
	ChapterSelectors []string `yaml:"chapter-selectors"`
	Pagination       `yaml:"pagination"`
	ChapterContent   `yaml:"chapter-content"`
}

// Chapter is the struct for a single chapter, requires on the URL
// also implements the ChapterContent struct
type Chapter struct {
	URL            string `yaml:"url"`
	ChapterContent `yaml:"chapter-content"`
}
