package config

// Replacement contains the replaced URL and their replacement
type Replacement struct {
	Url            string `yaml:"url"`
	ReplacementURL string `yaml:"replacement"`
}
