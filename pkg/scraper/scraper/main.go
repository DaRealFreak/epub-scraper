package scraper

import (
	"log"
	"os"
	"path/filepath"

	"github.com/DaRealFreak/epub-scraper/pkg/scraper/config"
	"github.com/bmaupin/go-epub"
)

// Scraper is the main functionality struct
type Scraper struct {
}

// NewScraper returns a new scraper struct
func NewScraper() *Scraper {
	return &Scraper{}
}

// HandleDirectory checks and handles all configurations in a passed directory
func (s *Scraper) HandleDirectory(directoryName string) {
	files, err := s.filePathWalkDir(directoryName)
	if err != nil {
		log.Fatal(err)
	}
	for _, filePath := range files {
		s.HandleFile(filePath)
	}
}

// HandleFile handles a single passed configuration file
func (s *Scraper) HandleFile(fileName string) {
	cfg, err := config.ReadConfigurationFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	s.WriteEPUB(cfg)
}

// filePathWalkDir use filepath.Walk to recursively retrieve all files from the passed directory
func (s *Scraper) filePathWalkDir(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// WriteEPUB writes the epub file and saves it in the main directory
func (s *Scraper) WriteEPUB(cfg *config.NovelConfig) {
	e := epub.NewEpub(cfg.General.Title)

	// Set the author
	e.SetAuthor(cfg.General.Author)

	// Add a section
	section := `<h1> Chapter x </h1>` +
		`<p>Example chapter</p>`
	_, err := e.AddSection(section, "Chapter x", "", "")

	// Write the EPUB
	err = e.Write(cfg.General.Title + ".epub")
	if err != nil {
		log.Fatal(err)
	}
}
