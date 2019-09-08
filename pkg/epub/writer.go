package epub

import (
	"fmt"
	"github.com/DaRealFreak/epub-scraper/pkg/config"
	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"github.com/bmaupin/go-epub"
	"path/filepath"
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
	writer.createEpub()
	writer.importAssets()
	writer.importAndAddCover()
	return writer
}

// createEpub creates epub writer and sets the initial variables from the cfg
func (w *Writer) createEpub() {
	w.Epub = epub.NewEpub(w.cfg.General.Title)

	// Set the author
	w.Epub.SetAuthor(w.cfg.General.Author)
}

// WriteEPUB writes the generated epub to the file system
func (w *Writer) WriteEPUB() {
	// save the .epub file to the drive
	raven.CheckError(w.Epub.Write(w.cfg.General.Title + ".epub"))
}

func (w *Writer) AddChapter(title string, content string) {
	section := fmt.Sprintf(`<h1> %s </h1>`+
		`%s`, title, content)
	_, _ = w.Epub.AddSection(section, title, "", w.cfg.Assets.Css.InternalPath)
}

// importAssets adds the specified assets to the epub
func (w *Writer) importAssets() {
	if w.cfg.Assets.Css.HostPath != "" {
		if !filepath.IsAbs(w.cfg.Assets.Css.HostPath) {
			// if not an absolute path we combine it with our configuration file bath
			w.cfg.Assets.Css.HostPath = filepath.Join(w.cfg.BaseDirectory, w.cfg.Assets.Css.HostPath)
		}
		internalPath, err := w.Epub.AddCSS(w.cfg.Assets.Css.HostPath, filepath.Base(w.cfg.Assets.Css.HostPath))
		raven.CheckError(err)
		w.cfg.Assets.Css.InternalPath = internalPath
	}
	if w.cfg.Assets.Font.HostPath != "" {
		if !filepath.IsAbs(w.cfg.Assets.Font.HostPath) {
			// if not an absolute path we combine it with our configuration file bath
			w.cfg.Assets.Font.HostPath = filepath.Join(w.cfg.BaseDirectory, w.cfg.Assets.Font.HostPath)
		}
		internalPath, err := w.Epub.AddFont(w.cfg.Assets.Font.HostPath, filepath.Base(w.cfg.Assets.Font.HostPath))
		raven.CheckError(err)
		w.cfg.Assets.Font.InternalPath = internalPath
	}
}

// importAndAddCover adds the specified cover to the epub
func (w *Writer) importAndAddCover() {
	// no need to add cover if no cover is set
	if w.cfg.General.Cover == "" {
		return
	}

	internalFilePath, err := w.Epub.AddImage(w.cfg.General.Cover, "cover"+filepath.Ext(w.cfg.General.Cover))
	raven.CheckError(err)

	section := fmt.Sprintf(`<img src="%s" height="100%%"/>`, internalFilePath)
	_, err = w.Epub.AddSection(section, "Cover", "", "")
	raven.CheckError(err)
}
