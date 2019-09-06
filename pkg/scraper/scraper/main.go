package scraper

import (
	"github.com/DaRealFreak/epub-scraper/pkg/scraper/config"
	"github.com/DaRealFreak/epub-scraper/pkg/scraper/epub"
	log "github.com/sirupsen/logrus"
)

// Scraper is the main functionality struct
type Scraper struct {
}

// NewScraper returns a new scraper struct
func NewScraper() *Scraper {
	return &Scraper{}
}

// HandleFile handles a single passed configuration file
func (s *Scraper) HandleFile(fileName string) {
	cfg, err := config.ReadConfigurationFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	writer := epub.NewWriter(cfg)

	// Add a section
	section := `<h1> Chapter x </h1>` +
		`<p>Example chapter</p>`
	_, err = writer.Epub.AddSection(section, "Chapter x", "", "")
	writer.WriteEPUB()
}
