package update

import (
	"fmt"
	"os"

	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"github.com/DaRealFreak/epub-scraper/pkg/version"
	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	log "github.com/sirupsen/logrus"
	"github.com/tcnksm/go-gitconfig"
)

// Checker is a struct solely to prevent expose functions without setting up first
type Checker struct {
}

// NewUpdateChecker returns a new update checker
func NewUpdateChecker() *Checker {
	return &Checker{}
}

// CheckForAvailableUpdates checks if any new releases exist and prints a notification if there is a new release
func (u *Checker) CheckForAvailableUpdates() {
	// check for available updates
	updateAvailable, err := u.isUpdateAvailable()
	raven.CheckError(err)
	if updateAvailable {
		fmt.Println("new version detected, run \"watcher update\" to update your application.")
	}
}

// isUpdateAvailable checks the latest release and compares it with the version of the run application
func (u *Checker) isUpdateAvailable() (updateAvailable bool, err error) {
	latest, found, err := selfupdate.DetectLatest(version.RepositoryURL)
	if err != nil {
		log.Warning("error occurred while detecting version: ", err)
		return false, err
	}

	v := semver.MustParse(version.VERSION)
	if !found || latest.Version.LTE(v) {
		return false, nil
	}
	return true, nil
}

// UpdateApplication downloads the update and directly updates the application
func (u *Checker) UpdateApplication() (err error) {
	updateAvailable, err := u.isUpdateAvailable()
	if err != nil {
		return err
	}
	if !updateAvailable {
		fmt.Println("current version is the latest")
		return nil
	}

	// check for github token in environment or in git config if not set in environment
	// required for updates when repository is private
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token, _ = gitconfig.GithubToken()
	}

	up, err := selfupdate.NewUpdater(selfupdate.Config{
		APIToken: token,
	})
	if err != nil {
		return err
	}

	fmt.Println("new version detected, updating...")
	// retrieve latest asset url again
	latest, _, err := up.DetectLatest(version.RepositoryURL)
	if err != nil {
		log.Warning("error occurred while retrieving latest asset URLs: ", err)
		return err
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not locate executable path")
	}
	if err := up.UpdateTo(latest, exe); err != nil {
		return err
	}
	log.Info("successfully updated to version " + latest.Version.String())
	return nil
}
