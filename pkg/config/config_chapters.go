package config

// Source is the option to define the source of the chapter content, table of content or single chapter
type Source struct {
	Toc     *Toc     `yaml:"toc"`
	Chapter *Chapter `yaml:"chapter"`
}

// SourceContent contains all configurations required for any type of source
type SourceContent struct {
	TitleContent   `yaml:"title-content"`
	ChapterContent `yaml:"chapter-content"`
}

// Toc is the struct for a table of content, requires the URL and the ChapterSelectors
// and implements the Pagination struct and the ChapterContent struct
type Toc struct {
	URL             string `yaml:"url"`
	ChapterSelector string `yaml:"chapter-selector"`
	Pagination      `yaml:"pagination"`
	SourceContent   `yaml:",inline"`
}

// Chapter is the struct for a single chapter, requires on the URL
// also implements the ChapterContent struct
type Chapter struct {
	URL           string `yaml:"url"`
	SourceContent `yaml:",inline"`
}
