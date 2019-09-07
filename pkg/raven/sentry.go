package raven

import (
	"io"
	"time"

	"github.com/DaRealFreak/epub-scraper/pkg/version"
	"github.com/getsentry/sentry-go"
	log "github.com/sirupsen/logrus"
)

// SetupSentry initializes the sentry
func SetupSentry() {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:     "https://b1eeaec0dada4b79bb8eb22be47a048a@sentry.io/1568260",
		Release: "epub-scraper@" + version.VERSION,
	}); err != nil {
		log.Fatal(err)
	}
}

// CheckError checks if the passed error is not nil and passes it to the sentry DSN
func CheckError(err error) {
	if err != nil {
		sentry.CaptureException(err)
		// Since sentry emits events in the background we need to make sure
		// they are sent before we shut down
		sentry.Flush(time.Second * 5)
		log.Fatal(err)
	}
}

// CheckClosure checks for errors on closeable objects
func CheckClosure(obj io.Closer) {
	CheckError(obj.Close())
}
