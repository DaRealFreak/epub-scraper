package scraper

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"github.com/DaRealFreak/epub-scraper/pkg/session"
	log "github.com/sirupsen/logrus"
)

// checkRedirect checks the passed request for token fragments and returns http.ErrUseLastResponse if found
// this causes the client to not follow the redirect, enabling us to use non-existing URLs as redirect URL
func (s *WaybackMachineWrapper) checkRedirect(req *http.Request, via []*http.Request) error {
	siteConfig := s.cfg.GetSiteConfigFromURL(req.URL)
	if siteConfig.WaybackMachine.Use {
		return &session.UseWaybackMachineError{
			Handling: &siteConfig.WaybackMachine,
			URL:      req.URL,
		}
	}
	log.Infof(req.URL.String())

	// return the previously set redirect function
	if s.sessionRedirect != nil {
		return s.sessionRedirect(req, via)
	}
	// session redirect can be nil too, fallback to default http.Client -> defaultCheckRedirect
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	return nil
}

// Get performs a normal GET request but checks the redirects to a host which should use the wayback machine
func (s *WaybackMachineWrapper) Get(uri string) (response *http.Response, err error) {
	// save original check redirect function and replace it with our custom one
	s.sessionRedirect = s.session.Client.CheckRedirect
	s.session.Client.CheckRedirect = s.checkRedirect

	// check direct passed URL for wayback machine host option and update url if required
	parsedURL, err := url.Parse(uri)
	raven.CheckError(err)
	siteConfig := s.cfg.GetSiteConfigFromURL(parsedURL)
	if siteConfig.WaybackMachine.Use {
		uri = fmt.Sprintf("https://web.archive.org/web/%s/%s", siteConfig.WaybackMachine.Version, uri)
	}

	// make the get request
	response, err = s.session.Get(uri)

	// if we previously updated the uri we restore the original request URL again for host settings
	if response != nil && siteConfig.WaybackMachine.Use {
		response.Request.URL = parsedURL
	}

	// restore original check redirect function
	s.session.Client.CheckRedirect = s.sessionRedirect
	return response, err
}
