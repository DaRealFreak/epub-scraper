package epub

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"os/exec"
	"path/filepath"

	"github.com/DaRealFreak/epub-scraper/pkg/config"
	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"github.com/DaRealFreak/epub-scraper/pkg/version"
	"github.com/bmaupin/go-epub"
	log "github.com/sirupsen/logrus"
)

// chapter contains all relevant of the added chapters
type chapter struct {
	title     string
	content   string
	addPrefix bool
}

// Writer contains all information and functions to create the final .epub file
type Writer struct {
	Epub     *epub.Epub
	chapters []chapter
	cfg      *config.NovelConfig
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

// createEpub creates epub writer and sets the available metadata taken from the configuration
func (w *Writer) createEpub() {
	w.Epub = epub.NewEpub(w.cfg.General.Title)

	// Set the meta data used for libraries in readers
	w.Epub.SetAuthor(w.cfg.General.Author)
	log.Infof("set author to: %s", w.cfg.General.Author)
	w.Epub.SetDescription(w.cfg.General.Description)
	log.Infof("set description to: %s", w.cfg.General.Description)
	w.Epub.SetLang(w.cfg.General.Language)
	log.Infof("set language to: %s", w.cfg.General.Language)
}

// WriteEpub writes the generated epub to the file system
func (w *Writer) WriteEpub() {
	w.createToC()
	w.writeChapters()
	// save the .epub file to the drive
	path, err := filepath.Abs(filepath.Clean(w.cfg.General.Title + ".epub"))
	raven.CheckError(err)
	raven.CheckError(w.Epub.Write(path))
	log.Infof("epub saved to %s", path)
}

// PolishEpub uses calibres ebook-polish command to compress images and fix possible errors
// which occurred to me multiple times using the bmaupin/go-epub library
func (w *Writer) PolishEpub() {
	path, err := filepath.Abs(filepath.Clean(w.cfg.General.Title + ".epub"))
	raven.CheckError(err)
	// #nosec
	raven.CheckError(exec.Command("ebook-polish", "-U", "-i", path, path).Run())
	log.Infof("generated epub got successfully polished")
}

// AddChapter adds a chapter to the to our current chapter list
func (w *Writer) AddChapter(title string, content string, addPrefix bool) {
	w.chapters = append(w.chapters, chapter{title: title, content: content, addPrefix: addPrefix})
}

// createToC creates a table of contents page to jump directly to chapters
// uses the previously appended chapters to link them
func (w *Writer) createToC() {
	t := template.Must(template.New("").Parse(`
		<div>
            <h3>{{.title}}</h3>
            {{.altTitle}}
			<div class="left">
				<p><a href="{{.rawUrl}}">Original Webnovel</a> by {{.author}}</p>
    			{{.toc}}
			</div>
			<div class="small-font bottom-align center">
				<p>Visit the translators at:<br/>
					{{.translators}}
				</p>
				<p>Epub created by: <br/>
					DaRealFreak <a href="https://github.com/{{.repositoryUrl}}">(Epub Creator Project)</a>
				</p>
			</div>
        </div>`))

	toc := ""
	for index, savedChapter := range w.chapters {
		chapterTitle := savedChapter.title
		// add prefix if requested (optional since many add it already in the ToC)
		if savedChapter.addPrefix {
			chapterTitle = fmt.Sprintf("Chapter %d - %s", index+1, chapterTitle)
		}
		toc += fmt.Sprintf(
			`<p><a href="chapter%04d.xhtml">%s</a></p>`,
			index+1,
			chapterTitle,
		)
	}

	// since alt title is optional we set it only if not empty
	altTitle := ""
	if w.cfg.General.AltTitle != "" {
		altTitle = fmt.Sprintf(`<h4><i>- %s -</i></h4>`, html.EscapeString(w.cfg.General.AltTitle))
	}

	translators := ""
	for _, translator := range w.cfg.General.Translators {
		translators += fmt.Sprintf(
			`<a href="%s">%s</a><br/>`, translator.URL, html.EscapeString(translator.Name),
		)
	}

	contentBuffer := new(bytes.Buffer)
	// #nosec
	raven.CheckError(t.Execute(contentBuffer, map[string]interface{}{
		"title":         w.cfg.General.Title,
		"altTitle":      template.HTML(altTitle),
		"rawUrl":        w.cfg.General.Raw,
		"author":        w.cfg.General.Author,
		"toc":           template.HTML(toc),
		"translators":   template.HTML(translators),
		"repositoryUrl": version.RepositoryURL,
	}))
	_, err := w.Epub.AddSection(
		contentBuffer.String(),
		"Table of Contents",
		"content.xhtml",
		w.cfg.Assets.CSS.InternalPath,
	)
	raven.CheckError(err)
}

// writeChapters writes all appended chapters to the epub file
func (w *Writer) writeChapters() {
	for index, savedChapter := range w.chapters {
		chapterTitle := savedChapter.title
		// add prefix if requested (optional since many add it already in the ToC)
		if savedChapter.addPrefix {
			chapterTitle = fmt.Sprintf("Chapter %d - %s", index+1, chapterTitle)
		}
		t := template.Must(template.New("").Parse(`
			<div class="left" style="text-align:left;text-indent:0;">
				<h3>{{.chapterTitle}}</h3>
				<hr/>
				{{.content}}
			</div>`))

		contentBuffer := new(bytes.Buffer)
		raven.CheckError(t.Execute(contentBuffer, map[string]interface{}{
			"chapterTitle": chapterTitle,
			// #nosec
			"content": template.HTML(savedChapter.content),
		}))

		_, err := w.Epub.AddSection(
			contentBuffer.String(),
			chapterTitle,
			fmt.Sprintf("chapter%04d.xhtml", index+1),
			w.cfg.Assets.CSS.InternalPath,
		)
		raven.CheckError(err)
	}
}

// importAssets adds the specified assets to the epub
func (w *Writer) importAssets() {
	if w.cfg.Assets.CSS.HostPath != "" {
		if !filepath.IsAbs(w.cfg.Assets.CSS.HostPath) {
			// if not an absolute path we combine it with our configuration file bath
			w.cfg.Assets.CSS.HostPath = filepath.Join(w.cfg.BaseDirectory, w.cfg.Assets.CSS.HostPath)
		}
		internalPath, err := w.Epub.AddCSS(w.cfg.Assets.CSS.HostPath, filepath.Base(w.cfg.Assets.CSS.HostPath))
		raven.CheckError(err)
		w.cfg.Assets.CSS.InternalPath = internalPath
		log.Infof("imported CSS file: %s", w.cfg.Assets.CSS.HostPath)
	}
	if w.cfg.Assets.Font.HostPath != "" {
		if !filepath.IsAbs(w.cfg.Assets.Font.HostPath) {
			// if not an absolute path we combine it with our configuration file bath
			w.cfg.Assets.Font.HostPath = filepath.Join(w.cfg.BaseDirectory, w.cfg.Assets.Font.HostPath)
		}
		internalPath, err := w.Epub.AddFont(w.cfg.Assets.Font.HostPath, filepath.Base(w.cfg.Assets.Font.HostPath))
		raven.CheckError(err)
		w.cfg.Assets.Font.InternalPath = internalPath
		log.Infof("imported font file: %s", w.cfg.Assets.Font.HostPath)
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

	w.Epub.SetCover(internalFilePath, w.cfg.Assets.CSS.InternalPath)
	log.Infof("set cover to: %s", w.cfg.General.Cover)
}
