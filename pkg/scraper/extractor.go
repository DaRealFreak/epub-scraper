package scraper

import "github.com/DaRealFreak/epub-scraper/pkg/config"

func (s *Scraper) handleToC(toc *config.Toc) {
	/*
		open URL
		if pagination {
			chapters = extract chapters from pagination
		} else {
			chapters = extract chapters from current URL
		}
		// option to reverse posts to always start at page 1 for config not to work just "in the moment"
		if pagination.ReversePosts {
			chapters = reverse chapters
		}
		foreach chapters {
			foreach ToC.ChapterSelectors {
				// ignore if selector not found -> option for links to blog posts instead of direct chapter to be skipped
				if selector found {
					content = follow link of chapter selector
				}
			}
			if content {
				handleChapterContent(content)
			} else {
				fatal error, no content extractable
			}
		}
	*/
}

func (s *Scraper) handleChapter(chapter *config.Chapter) {
	/*
		content = extract content from chapter.URL
		handleChapterContent(content)
	*/
}
