package main

import (
	"os"

	"github.com/DaRealFreak/epub-scraper/pkg/scraper/scraper"
	log "github.com/sirupsen/logrus"
)

func main() {
	app := scraper.NewScraper()
	for _, s := range os.Args[1:] {
		if _, err := os.Stat(s); os.IsNotExist(err) {
			log.Fatalf("%s is neither a file or directory", s)
		}

		// #nosec
		file, err := os.Open(s)
		if err != nil {
			log.Fatal(err)
		}

		fi, err := file.Stat()
		if err != nil {
			log.Fatal(err)
		}

		err = file.Close()
		switch {
		case err != nil:
			log.Fatal(err)
		case fi.IsDir():
			app.HandleDirectory(s)
		default:
			app.HandleFile(s)
		}
	}
}
