package scraper

import (
	"os"

	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"github.com/DaRealFreak/epub-scraper/pkg/scraper"
	"github.com/DaRealFreak/epub-scraper/pkg/update"
	"github.com/DaRealFreak/epub-scraper/pkg/version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Scraper returns the CLI Scraper struct
type Scraper struct {
	logLevel string
	rootCmd  *cobra.Command
}

// NewScraper returns the pointer to an initialized CLI Scraper struct
func NewScraper() *Scraper {
	app := &Scraper{
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
			},
		},
	}

	app.rootCmd.PersistentFlags().StringVarP(
		&app.logLevel,
		"verbosity",
		"v",
		log.InfoLevel.String(),
		"log level (debug, info, warn, error, fatal, panic)",
	)

	// add sub commands
	app.addUpdateCommand()

	// parse all configurations before executing the main command
	cobra.OnInitialize(app.initScraper)
	return app
}

// Execute executes the root command, entry point for the CLI application
func (cli *Scraper) Execute() {
	// check for available updates
	update.NewUpdateChecker().CheckForAvailableUpdates()

	if err := cli.rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

// initScraper initializes everything the CLI application needs
func (cli *Scraper) initScraper() {
	// setup sentry for error logging
	raven.SetupSentry()

	// set log level
	lvl, err := log.ParseLevel(cli.logLevel)
	raven.CheckError(err)
	log.SetLevel(lvl)
}

// addUpdateCommand adds the update sub command
func (cli *Scraper) addUpdateCommand() {
	// general add option
	addCmd := &cobra.Command{
		Use:   "update",
		Short: "update the application",
		Long:  "function for the user to update the application",
		Run: func(cmd *cobra.Command, args []string) {
			err := update.NewUpdateChecker().UpdateApplication()
			raven.CheckError(err)
		},
	}
	cli.rootCmd.AddCommand(addCmd)
}
