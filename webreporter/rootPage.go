package webreporter

import (
	"net/http"
	"time"
)

func (obj *webReporter) rootPage(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		obj.logger.Printf("NotFound: %s %s", req.Method, req.URL.Path)
		http.NotFound(w, req)
		return
	}

	mainURL := "/display"

	// Perform the 302 redirect
	// http.StatusFound is the constant for 302 Found
	http.Redirect(w, req, mainURL, http.StatusFound)
}

///////////////////////////////////////////////////////////////////////////////

type rootDetails struct {
	Title, Version                                string
	ProcessingSize, ProcessingSpeed               int64
	ProcessingTime, FirstEventTime, LastEventTime time.Time
}

func (obj *webReporter) getRootDetails() (data rootDetails) {

	details := obj.storage.SelectQuery("details")
	details.Next(
		&data.Title, &data.Version,
		&data.ProcessingSize, &data.ProcessingSpeed,
		&data.ProcessingTime,
		&data.FirstEventTime, &data.LastEventTime)

	details.Next()

	return
}
