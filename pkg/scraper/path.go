package scraper

import (
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// HandleDirectory checks and handles all configurations in a passed directory
func (s *Scraper) HandleDirectory(directoryName string) {
	files, err := s.filePathWalkDir(directoryName)
	if err != nil {
		log.Fatal(err)
	}
	for _, filePath := range files {
		if strings.HasSuffix(filePath, ".yaml") {
			s.HandleFile(filePath)
		}
	}
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
