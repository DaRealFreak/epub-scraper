package epub

import (
	"bytes"
	"fmt"
	"html"
	"html/template"

	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"github.com/DaRealFreak/epub-scraper/pkg/version"
)

// getChapterTitle returns the chapter title parsed with the configured template
func (w *Writer) getChapterTitle(savedChapter chapter, chapterIndex int) string {
	chapterTitle := savedChapter.title
	// add prefix if requested (optional since many add it already in the ToC)
	if savedChapter.addPrefix {
		if w.cfg.Templates.Chapter.Title == "" {
			w.cfg.Templates.Chapter.Title = `Chapter {{.chapterIndex}} - {{.chapterTitle}}`
		}
		chapterTemplate := template.Must(template.New("").Parse(w.cfg.Templates.Chapter.Title))
		buffer := new(bytes.Buffer)
		raven.CheckError(chapterTemplate.Execute(buffer, map[string]interface{}{
			"chapterIndex": chapterIndex + 1,
			"chapterTitle": chapterTitle,
		}))
		chapterTitle = buffer.String()
	}
	return chapterTitle
}

// getTranslators returns the translator list parsed with the configured template
func (w *Writer) getTranslators() string {
	if w.cfg.Templates.ToC.Translator == "" {
		w.cfg.Templates.ToC.Translator = `<a href="{{.translatorURL}}">{{.translatorName}}</a><br/>`
	}
	translatorTemplate := template.Must(template.New("").Parse(w.cfg.Templates.ToC.Translator))
	translators := ""
	for _, translator := range w.cfg.General.Translators {
		buffer := new(bytes.Buffer)
		raven.CheckError(translatorTemplate.Execute(buffer, map[string]interface{}{
			"translatorURL":  translator.URL,
			"translatorName": translator.Name,
		}))
		translators += buffer.String()
	}
	return translators
}

// getAltTitle returns the parsed template of the optional alt title
// if no alt title is defined it'll return an empty string
func (w *Writer) getAltTitle() string {
	// since alt title is optional we set it only if not empty
	altTitle := ""
	if w.cfg.General.AltTitle != "" {
		if w.cfg.Templates.AltTitle == "" {
			w.cfg.Templates.AltTitle = `<h4><i>- {{.altTitle}} -</i></h4>`
		}
		altTitleTemplate := template.Must(template.New("").Parse(w.cfg.Templates.AltTitle))
		buffer := new(bytes.Buffer)
		raven.CheckError(altTitleTemplate.Execute(buffer, map[string]interface{}{
			"altTitle": html.EscapeString(w.cfg.General.AltTitle),
		}))
		altTitle = buffer.String()
	}
	return altTitle
}

// getToC returns a table of contents consisting of a simple list of links to the chapter with the chapter title as name
func (w *Writer) getToC() string {
	toc := ""
	for index, savedChapter := range w.chapters {
		chapterTitle := w.getChapterTitle(savedChapter, index)
		toc += fmt.Sprintf(
			`<p><a href="chapter%04d.xhtml">%s</a></p>`,
			index+1,
			chapterTitle,
		)
	}
	return toc
}

// getEpubScraperCredits returns the epub scraper credits including a link to the repository
func (w *Writer) getEpubScraperCredits() string {
	return fmt.Sprintf(`Epub created by: <br/>
		DaRealFreak <a href="https://github.com/%s">(Epub Creator Project)</a>`, version.RepositoryURL)
}
