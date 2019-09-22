package session

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/DaRealFreak/epub-scraper/pkg/config"
	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	log "github.com/sirupsen/logrus"
)

// WaybackMachineWrapper contains wayback machine related variables like the session and the novel configuration
type WaybackMachineWrapper struct {
	session
	cfg             *config.NovelConfig
	sessionRedirect func(req *http.Request, via []*http.Request) error
}

// checkRedirect checks the passed request for token fragments and returns http.ErrUseLastResponse if found
// this causes the client to not follow the redirect, enabling us to use non-existing URLs as redirect URL
func (w *WaybackMachineWrapper) checkRedirect(req *http.Request, via []*http.Request) error {
	siteConfig := w.cfg.GetSiteConfigFromURL(req.URL)
	if siteConfig.WaybackMachine.Use {
		return &UseWaybackMachineError{
			Handling: &siteConfig.WaybackMachine,
			URL:      req.URL,
		}
	}
	log.Infof(req.URL.String())

	// return the previously set redirect function
	if w.sessionRedirect != nil {
		return w.sessionRedirect(req, via)
	}
	// session redirect can be nil too, fallback to default http.Client -> defaultCheckRedirect
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	return nil
}

// Get performs a normal GET request but checks the redirects to a host which should use the wayback machine
func (w *WaybackMachineWrapper) Get(uri string) (response *http.Response, err error) {
	// save original check redirect function and replace it with our custom one
	w.sessionRedirect = w.session.Client.CheckRedirect
	w.session.Client.CheckRedirect = w.checkRedirect

	// check direct passed URL for wayback machine host option and update url if required
	parsedURL, err := url.Parse(uri)
	raven.CheckError(err)
	siteConfig := w.cfg.GetSiteConfigFromURL(parsedURL)
	if siteConfig.WaybackMachine.Use {
		uri = fmt.Sprintf("https://web.archive.org/web/%s/%s", siteConfig.WaybackMachine.Version, uri)
	}

	// make the get request
	response, err = w.session.Get(uri)

	// if we previously updated the uri we restore the original request URL again for host settings
	if response != nil && siteConfig.WaybackMachine.Use {
		response.Request.URL = parsedURL
	}

	// restore original check redirect function
	w.session.Client.CheckRedirect = w.sessionRedirect
	return response, err
}
