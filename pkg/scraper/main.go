package scraper

import (
	"github.com/DaRealFreak/epub-scraper/pkg/config"
	"github.com/DaRealFreak/epub-scraper/pkg/epub"
	"github.com/DaRealFreak/epub-scraper/pkg/session"
	log "github.com/sirupsen/logrus"
)

// Scraper is the main functionality struct
type Scraper struct {
	session *session.Session
}

// NewScraper returns a new scraper struct
func NewScraper() *Scraper {
	return &Scraper{
		session: session.NewSession(),
	}
}

// HandleFile handles a single passed configuration file
func (s *Scraper) HandleFile(fileName string) {
	cfg, err := config.ReadConfigurationFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	writer := epub.NewWriter(cfg)
	writer.AddChapter("Ruti is cute", "Hello world!")
	writer.WriteEPUB()
}
