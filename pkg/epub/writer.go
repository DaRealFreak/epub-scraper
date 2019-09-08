package epub

import (
	"bytes"
	"fmt"
	"github.com/DaRealFreak/epub-scraper/pkg/config"
	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"github.com/DaRealFreak/epub-scraper/pkg/version"
	"github.com/bmaupin/go-epub"
	"html"
	"html/template"
	"path/filepath"
)

type chapter struct {
	title   string
	content string
}

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

// createEpub creates epub writer and sets the initial variables from the cfg
func (w *Writer) createEpub() {
	w.Epub = epub.NewEpub(w.cfg.General.Title)

	// Set the author
	w.Epub.SetAuthor(w.cfg.General.Author)
}

// WriteEPUB writes the generated epub to the file system
func (w *Writer) WriteEPUB() {
	w.createToC()
	w.writeChapters()
	// save the .epub file to the drive
	raven.CheckError(w.Epub.Write(w.cfg.General.Title + ".epub"))
}

// AddChapter adds a chapter to the to our current list
func (w *Writer) AddChapter(title string, content string) {
	w.chapters = append(w.chapters, chapter{title: title, content: content})
}

// createToC creates a table of contents page to jump directly to chapters
// uses the previously appended chapters to link them
func (w *Writer) createToC() {
	t := template.Must(template.New("").Parse(`
		<div>
            <h3>{{.title}}</h3>
            {{.altTitle}}
			<div class="center">
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
		toc += fmt.Sprintf(
			`<p><a href="chapter%04d.xhtml">%s</a></p>`,
			index+1,
			fmt.Sprintf("Chapter %d - %s", index+1, savedChapter.title),
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
			`<a href="%s">%s</a><br/>`, translator.Url, html.EscapeString(translator.Name),
		)
	}

	contentBuffer := new(bytes.Buffer)
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
		string(contentBuffer.Bytes()),
		"Table of Contents",
		"content.xhtml",
		w.cfg.Assets.Css.InternalPath,
	)
	raven.CheckError(err)
}

// writeChapters writes all appended chapters to the epub file
func (w *Writer) writeChapters() {
	for index, savedChapter := range w.chapters {
		chapterTitle := fmt.Sprintf("Chapter %d - %s", index+1, savedChapter.title)
		t := template.Must(template.New("").Parse(`
			<div class="left" style="text-align:left;text-indent:0;">
				<h3>{{.chapterTitle}}</h3>
				<hr/>
				{{.content}}
			</div>`))

		contentBuffer := new(bytes.Buffer)
		raven.CheckError(t.Execute(contentBuffer, map[string]interface{}{
			"chapterTitle": chapterTitle,
			"content":      template.HTML(savedChapter.content),
		}))

		_, err := w.Epub.AddSection(
			string(contentBuffer.Bytes()),
			chapterTitle,
			fmt.Sprintf("chapter%04d.xhtml", index+1),
			w.cfg.Assets.Css.InternalPath,
		)
		raven.CheckError(err)
	}
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

	section := fmt.Sprintf(`<img src="%s" alt="Cover Image"/>`, internalFilePath)
	_, err = w.Epub.AddSection(section, "Cover", "cover.xhtml", w.cfg.Assets.Css.InternalPath)
	raven.CheckError(err)
}
