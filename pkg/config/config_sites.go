package config

// SiteConfiguration is an optional configuration to extract the Pagination struct and the ChapterContent struct
// into one Configuration object, allowing multiple sources of the same Host to reuse them
// Source options have a higher priority than the SiteConfiguration options
type SiteConfiguration struct {
	Host           string         `yaml:"host"`
	Pagination     Pagination     `yaml:"pagination"`
	TitleContent   TitleContent   `yaml:"title-content"`
	ChapterContent ChapterContent `yaml:"chapter-content"`
}
