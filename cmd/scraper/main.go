package scraper

import (
	"fmt"
	"log"
	"os"

	"github.com/DaRealFreak/epub-scraper/pkg/scraper"
	"github.com/DaRealFreak/epub-scraper/pkg/update"
	"github.com/DaRealFreak/epub-scraper/pkg/version"
	"github.com/spf13/cobra"
)

// Scraper returns the CLI Scraper struct
type Scraper struct {
	rootCmd *cobra.Command
}

// NewScraper returns the pointer to an initialized CLI Scraper struct
func NewScraper() *Scraper {
	return &Scraper{
		rootCmd: &cobra.Command{
			Use:   "scraper",
			Short: "Scraper scraps novels from websites and generates a ready to read .epub file.",
			Long: "An application written in Go to scrap novel chapters from websites to create an .epub file.\n" +
				"You can pass a configuration file with lots of configuration options " +
				"to work with as many websites as possible",
			Version: version.VERSION,
			Args:    cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				app := scraper.NewScraper()
				for _, s := range args {
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

				fmt.Print(args)
			},
		},
	}
}

// Execute executes the root command, entry point for the CLI application
func (cli *Scraper) Execute() {
	// check for available updates
	update.NewUpdateChecker().CheckForAvailableUpdates()

	if err := cli.rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
