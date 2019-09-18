package config

// Templates contains a collection of templates to style the generated epub file
type Templates struct {
	ToC     TemplateToC     `yaml:"toc"`
	Chapter TemplateChapter `yaml:"chapter"`
}

// TemplateToC contains all templates related to the table of content page
type TemplateToC struct {
	Content    string `yaml:"content"`
	AltTitle   string `yaml:"alt-title"`
	Translator string `yaml:"translator"`
}

// TemplateChapter contains all templates related to the chapter pages
type TemplateChapter struct {
	Content string `yaml:"content"`
	Title   string `yaml:"title"`
}
