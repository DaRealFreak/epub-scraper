package scraper

import (
	"bytes"
	"github.com/DaRealFreak/emoji-sanitizer/pkg/sanitizer/options"
	"strings"

	"github.com/DaRealFreak/emoji-sanitizer/pkg/sanitizer"
	"github.com/DaRealFreak/epub-scraper/pkg/config"
	"github.com/DaRealFreak/epub-scraper/pkg/epub"
	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"github.com/DaRealFreak/epub-scraper/pkg/session"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

// Scraper is the main functionality struct
type Scraper struct {
	configParser *config.Parser
	sanitizer    *sanitizer.Sanitizer
	session      session.Session
}

// ChapterData contains all relevant chapter data for writing them into the epub
type ChapterData struct {
	addPrefix bool
	title     string
	content   string
}

// NewScraper returns a new scraper struct
func NewScraper() (_ *Scraper, err error) {
	scraper := &Scraper{
		configParser: config.NewParser(),
	}
	scraper.sanitizer, err = sanitizer.NewSanitizer(
		options.UnicodeVersion(sanitizer.Version131),
		options.LoadFromOnline(true),
		options.UseFallbackToOffline(true),
		// allow common emojis implemented everywhere: "#", "*", "[0-9]", "©", "®", "‼", "™"
		options.AllowEmojiCodes([]string{"0023", "002A", "0030..0039", "00A9", "00AE", "203C", "2122"}),
	)

	return scraper, err
}

// HandleFile handles a single passed configuration file
func (s *Scraper) HandleFile(fileName string) {
	cfg, err := s.configParser.ReadConfigurationFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	s.session = session.NewSession(cfg)

	writer := epub.NewWriter(cfg)
	for _, source := range cfg.Chapters {
		if source.Toc != nil {
			chapters := s.handleToc(source.Toc, cfg)
			for _, chapter := range chapters {
				writer.AddChapter(chapter.title, chapter.content, chapter.addPrefix)
			}
		} else if source.Chapter != nil {
			chapter := s.extractChapterData(
				source.Chapter.URL,
				cfg,
				source.Chapter.SourceContent,
			)
			if chapter != nil {
				writer.AddChapter(chapter.title, chapter.content, chapter.addPrefix)
			}
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
