package scraper

import (
	"bytes"
	"strings"

	"github.com/DaRealFreak/epub-scraper/pkg/config"
	"github.com/DaRealFreak/epub-scraper/pkg/epub"
	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"github.com/DaRealFreak/epub-scraper/pkg/session"
	"github.com/microcosm-cc/bluemonday"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

// Scraper is the main functionality struct
type Scraper struct {
	configParser *config.Parser
	session      *session.Session
	sanitizer    *bluemonday.Policy
}

// NewScraper returns a new scraper struct
func NewScraper() *Scraper {
	return &Scraper{
		configParser: config.NewParser(),
		session:      session.NewSession(),
		sanitizer:    bluemonday.UGCPolicy(),
	}
}

// HandleFile handles a single passed configuration file
func (s *Scraper) HandleFile(fileName string) {
	cfg, err := s.configParser.ReadConfigurationFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	writer := epub.NewWriter(cfg)
	for _, source := range cfg.Chapters {
		if source.Toc != nil {
			chapters := s.handleToc(source.Toc)
			for _, chapter := range chapters {
				writer.AddChapter(chapter.title, chapter.content, chapter.addPrefix)
			}
		} else if source.Chapter != nil {
			chapter := s.handleChapter(source.Chapter)
			writer.AddChapter(chapter.title, chapter.content, chapter.addPrefix)
		}
	}
	// finally save the generated epub to the file system
	writer.WriteEpub()
	writer.PolishEpub()
}

// fixHTMLCode uses the net/html library to render the broken HTML code which mostly fixes broken HTML
func (s *Scraper) fixHTMLCode(htmlCode string) string {
	reader := strings.NewReader(htmlCode)
	root, err := html.Parse(reader)
	raven.CheckError(err)

	var b bytes.Buffer
	raven.CheckError(html.Render(&b, root))
	return b.String()
}
