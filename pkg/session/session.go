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

	"github.com/DaRealFreak/epub-scraper/pkg/raven"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

// Session is an extension to the implemented SessionInterface for HTTP sessions
type Session struct {
	Client     *http.Client
	MaxRetries int
	ctx        context.Context
}

// NewSession initializes a new session and sets all the required headers etc
func NewSession() *Session {
	jar, _ := cookiejar.New(nil)

	app := Session{
		Client:     &http.Client{Jar: jar},
		MaxRetries: 5,
		ctx:        context.Background(),
	}
	return &app
}

// Get sends a GET request, returns the occurred error if something went wrong even after multiple tries
func (s *Session) Get(uri string) (response *http.Response, err error) {
	// access the passed url and return the data or the error which persisted multiple retries
	// post the request with the retries option
	for try := 1; try <= s.MaxRetries; try++ {
		log.Debug(fmt.Sprintf("opening GET uri \"%s\" (try: %d)", uri, try))
		response, err = s.Client.Get(uri)
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

// Post sends a POST request, returns the occurred error if something went wrong even after multiple tries
func (s *Session) Post(uri string, data url.Values) (response *http.Response, err error) {
	// post the request with the retries option
	for try := 1; try <= s.MaxRetries; try++ {
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
func (s *Session) GetDocument(response *http.Response) *goquery.Document {
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
