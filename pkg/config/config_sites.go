package config

type SiteConfiguration struct {
	DNS            string         `yaml:"dns"`
	Pagination     Pagination     `yaml:"pagination"`
	ChapterContent ChapterContent `yaml:"chapter-content"`
}
