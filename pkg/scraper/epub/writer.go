package epub

import (
	"github.com/DaRealFreak/epub-scraper/pkg/scraper/config"
	"github.com/bmaupin/go-epub"
	log "github.com/sirupsen/logrus"
)

type Writer struct {
	Epub *epub.Epub
	cfg  *config.NovelConfig
}

// NewWriter returns a Writer struct
func NewWriter(cfg *config.NovelConfig) *Writer {
	writer := &Writer{
		cfg: cfg,
	}
	writer.createEPUB()
	return writer
}

// createEPUB creates epub writer and sets the initial variables from the cfg
func (w *Writer) createEPUB() {
	w.Epub = epub.NewEpub(w.cfg.General.Title)

	// Set the author
	w.Epub.SetAuthor(w.cfg.General.Author)
}

// WriteEPUB writes the generated epub to the file system
func (w *Writer) WriteEPUB() {
	// Write the EPUB
	if err := w.Epub.Write(w.cfg.General.Title + ".epub"); err != nil {
		log.Fatal(err)
	}
}
