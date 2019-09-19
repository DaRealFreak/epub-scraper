package epub

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"math/rand"
	"mime"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/DaRealFreak/epub-scraper/pkg/config"
	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"github.com/PuerkitoBio/goquery"
	"github.com/bmaupin/go-epub"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
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
	// rate limiter for importing assets
	RateLimiter *rate.Limiter
	ctx         context.Context
}

// NewWriter returns a Writer struct
func NewWriter(cfg *config.NovelConfig) *Writer {
	writer := &Writer{
		cfg:         cfg,
		RateLimiter: rate.NewLimiter(rate.Every(1500*time.Millisecond), 1),
		ctx:         context.Background(),
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
	w.extractAndImportImages(&content, len(w.chapters)+1)
	w.chapters = append(w.chapters, chapter{title: title, content: content, addPrefix: addPrefix})
}

// createToC creates a table of contents page to jump directly to chapters
// uses the previously appended chapters to link them
func (w *Writer) createToC() {
	if w.cfg.Templates.ToC.Content == "" {
		w.cfg.Templates.ToC.Content = `
		<div>
            <h3>{{.title}}</h3>
            {{.altTitle}}
			<div class="center">
				<p><a href="{{.rawUrl}}">Original Webnovel</a> by {{.author}}</p>
			</div>
			<div class="small-font bottom-align center">
				<p>Visit the translators at:<br/>
					{{.translators}}
				</p>
				<p>
					{{.epubScraperCredits}}
				</p>
			</div>
        </div>`
	}
	t := template.Must(template.New("").Parse(w.cfg.Templates.ToC.Content))

	toc := w.getToC()

	contentBuffer := new(bytes.Buffer)
	// #nosec
	raven.CheckError(t.Execute(contentBuffer, map[string]interface{}{
		"title":              w.cfg.General.Title,
		"altTitle":           template.HTML(w.getAltTitle()),
		"rawUrl":             w.cfg.General.Raw,
		"author":             w.cfg.General.Author,
		"toc":                template.HTML(toc),
		"translators":        template.HTML(w.getTranslators()),
		"epubScraperCredits": template.HTML(w.getEpubScraperCredits()),
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
		chapterTitle := w.getChapterTitle(savedChapter, index)
		if w.cfg.Templates.Chapter.Content == "" {
			w.cfg.Templates.Chapter.Content = `
				<div class="left" style="text-align:left;text-indent:0;">
					<h3>{{.chapterTitle}}</h3>
					<hr/>
					{{.content}}
				</div>`
		}
		t := template.Must(template.New("").Parse(w.cfg.Templates.Chapter.Content))

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

// extractAndImportImages extracts all external images, imports them into the epub and updates the display links
func (w *Writer) extractAndImportImages(content *string, chapterIndex int) {
	log.Infof("importing external assets into epub for chapter %d", chapterIndex)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(*content))
	raven.CheckError(err)

	doc.Find("img[src]").Each(func(i int, selection *goquery.Selection) {
		w.importValidMimeTypeFiles(content, chapterIndex, selection, "src")
	})

	doc.Find("[href]").Each(func(i int, selection *goquery.Selection) {
		w.importValidMimeTypeFiles(content, chapterIndex, selection, "href")
	})
}

// importValidMimeTypeFiles checks for the mime type by the file extension
// and imports them in case they match the allowed mime types for epub files which are documented here:
// https://www.w3.org/publishing/epub/epub-spec.html#sec-cmt-supported
func (w *Writer) importValidMimeTypeFiles(content *string, index int, selection *goquery.Selection, attrName string) {
	link, _ := selection.Attr(attrName)
	switch mime.TypeByExtension(filepath.Ext(link)) {
	case "image/gif", "image/jpeg", "image/png", "image/svg+xml":
		log.Debugf("importing external resource %s for chapter index %d", link, index)
		w.applyRateLimit()

		// retrieve the outer HTML for later replacement
		tag, err := goquery.OuterHtml(selection)
		raven.CheckError(err)

		// generate a unique file name and import the img source into the epub
		filename := fmt.Sprintf("%d_%s_%d%s",
			index,
			strings.TrimSuffix(filepath.Base(link), filepath.Ext(link)),
			// small random integer in case the same chapter links to multiple equal named images
			// no need for seeding, chance that we get a match is really low anyways
			// #nosec
			rand.Int(),
			filepath.Ext(link),
		)
		internalName, err := w.Epub.AddImage(link, filename)
		raven.CheckError(err)

		// update the src to our new internal file name
		selection.SetAttr(attrName, internalName)
		updatedTag, err := goquery.OuterHtml(selection)
		raven.CheckError(err)

		// update our content
		*content = strings.ReplaceAll(*content, tag, updatedTag)
	}
}

// applyRateLimit waits for the leaky bucket to fill again
func (w *Writer) applyRateLimit() {
	// if no rate limiter is defined we don't have to wait
	if w.RateLimiter != nil {
		// wait for request to stay within the rate limit
		err := w.RateLimiter.Wait(w.ctx)
		raven.CheckError(err)
	}
}
