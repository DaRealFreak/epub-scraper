package config

type Source struct {
	Toc     *Toc     `yaml:"toc"`
	Chapter *Chapter `yaml:"chapter"`
}

type Toc struct {
	URL              string   `yaml:"url"`
	ChapterSelectors []string `yaml:"chapter-selectors"`
	*Pagination      `yaml:"pagination"`
	*ChapterContent  `yaml:"chapter-content"`
}

type Chapter struct {
	URL             string `yaml:"url"`
	*ChapterContent `yaml:"chapter-content"`
}
