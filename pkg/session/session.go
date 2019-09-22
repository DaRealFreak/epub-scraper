package session

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/DaRealFreak/epub-scraper/pkg/config"
	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// Session is the interface for the implemented HTTP client
type Session interface {
	Get(uri string) (response *http.Response, err error)
	Post(uri string, data url.Values) (response *http.Response, err error)
	GetDocument(response *http.Response) *goquery.Document
	ApplyRateLimit()
}

// Session is an extension to the implemented SessionInterface for HTTP sessions
type session struct {
	Client      *http.Client
	RateLimiter *rate.Limiter
	MaxRetries  int
	ctx         context.Context
}

// UseWaybackMachineError custom error if we get redirected on a URL configured to use the wayback machine
type UseWaybackMachineError struct {
	error
	Handling *config.WaybackMachine
	URL      *url.URL
}

// NewSession initializes a new session and sets all the required headers etc
func NewSession(novelConfig *config.NovelConfig) Session {
	jar, _ := cookiejar.New(nil)

	app := session{
		Client:      &http.Client{Jar: jar},
		RateLimiter: rate.NewLimiter(rate.Every(1500*time.Millisecond), 1),
		MaxRetries:  5,
		ctx:         context.Background(),
	}

	return &WaybackMachineWrapper{
		session: app,
		cfg:     novelConfig,
	}
}

// Get sends a GET request, returns the occurred error if something went wrong even after multiple tries
func (s *session) Get(uri string) (response *http.Response, err error) {
	// access the passed url and return the data or the error which persisted multiple retries
	// post the request with the retries option
	for try := 1; try <= s.MaxRetries; try++ {
		s.ApplyRateLimit()
		log.Debug(fmt.Sprintf("opening GET uri \"%s\" (try: %d)", uri, try))
		response, err = s.Client.Get(uri)
		if err == nil && response.StatusCode < 400 {
			// if no error occurred and status code is okay too break out of the loop
			// 4xx & 5xx are client/server error codes, so we check for < 400
			return response, err
		}

		if waybackResponse, done, err := s.handleWaybackMachineError(response, err); done {
			return waybackResponse, err
		}

		// any other error falls into the retry clause
		time.Sleep(time.Duration(try+1) * time.Second)
	}
	return response, err
}

// handleWaybackMachineError checks if the returned error is indicating that we should use the wayback machine
// if yes we return the request using the wayback machine and replace the request URL to the original URL
// to keep host settings
func (s *session) handleWaybackMachineError(response *http.Response, err error) (*http.Response, bool, error) {
	if response != nil && err != nil {
		// can't use .(type) outside of switch case, so we have to use single case switch case here
		// nolint: gocritic
		switch v := err.(type) {
		case *url.Error:
			switch c := v.Err.(type) {
			case *UseWaybackMachineError:
				newURL := fmt.Sprintf("https://web.archive.org/web/%s/%s", c.Handling.Version, c.URL.String())
				newRes, err := s.Get(newURL)
				if newRes != nil {
					newRes.Request.URL = c.URL
					return newRes, true, err
				}
			}
		}
	}
	return nil, false, nil
}

// Post sends a POST request, returns the occurred error if something went wrong even after multiple tries
func (s *session) Post(uri string, data url.Values) (response *http.Response, err error) {
	// post the request with the retries option
	for try := 1; try <= s.MaxRetries; try++ {
		s.ApplyRateLimit()
		log.Debug(fmt.Sprintf("opening POST uri \"%s\" (try: %d)", uri, try))
		response, err = s.Client.PostForm(uri, data)
		switch {
		case err == nil && response.StatusCode < 400:
			// if no error occurred and status code is okay too break out of the loop
			// 4xx & 5xx are client/server error codes, so we check for < 400
			return response, err
		default:
			// any other error falls into the retry clause
			time.Sleep(time.Duration(try+1) * time.Second)
		}
	}
	return response, err
}

// GetDocument converts the http response to a *goquery.Document
func (s *session) GetDocument(response *http.Response) *goquery.Document {
	var reader io.ReadCloser
	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ = gzip.NewReader(response.Body)
	default:
		reader = response.Body
	}
	defer raven.CheckClosure(reader)
	document, err := goquery.NewDocumentFromReader(reader)
	raven.CheckError(err)
	return document
}

// ApplyRateLimit waits for the leaky bucket to fill again
func (s *session) ApplyRateLimit() {
	// if no rate limiter is defined we don't have to wait
	if s.RateLimiter != nil {
		// wait for request to stay within the rate limit
		err := s.RateLimiter.Wait(s.ctx)
		raven.CheckError(err)
	}
}
