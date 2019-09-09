package scraper

import (
	"github.com/DaRealFreak/epub-scraper/pkg/config"
	"github.com/DaRealFreak/epub-scraper/pkg/epub"
	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"github.com/DaRealFreak/epub-scraper/pkg/session"
	"github.com/PuerkitoBio/goquery"
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

	log.SetLevel(log.DebugLevel)

	// iterate through every ToC URL and append the extracted chapters
	for _, tocURL := range cfg.Toc.URLs {
		res, err := s.session.Get(tocURL.URL)
		raven.CheckError(err)
		doc := s.session.GetDocument(res)
		doc.Find(tocURL.ChapterSelector).Each(func(i int, selection *goquery.Selection) {
			chapterURL, exists := selection.Attr("href")
			if !exists {
				log.Warning("no chapter URL found")
			}
			writer.AddChapter(selection.Text(), s.getChapterContent(chapterURL, tocURL.ChapterContent))
		})
	}
	writer.WriteEPUB()
}
