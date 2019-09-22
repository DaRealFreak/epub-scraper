package config

import "encoding/json"

// SiteConfiguration is an optional configuration to extract the Pagination struct and the ChapterContent struct
// into one Configuration object, allowing multiple sources of the same Host to reuse them
// Source options have a higher priority than the SiteConfiguration options
type SiteConfiguration struct {
	Host           string     `yaml:"host"`
	Pagination     Pagination `yaml:"pagination"`
	SourceContent  `yaml:",inline"`
	Redirects      []string       `yaml:"redirects"`
	WaybackMachine WaybackMachine `yaml:"wayback-machine"`
}

// WaybackMachine contains the usage and version option of a site
type WaybackMachine struct {
	Use     bool        `yaml:"use"`
	Version json.Number `yaml:"version"`
}
