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

func (obj *webReporter) getDetails() {
	{
		data := obj.storage.Select("details")
		data.Next(
			&obj.details.Title, &obj.details.Version,
			// &data.ProcessingSize, &data.ProcessingSpeed,
			// &data.ProcessingTime,
			&obj.details.FirstEventTime, &obj.details.LastEventTime)

		data.Next()
	}
	{ // process counters
		var id int
		var name string

		data := obj.storage.Select("processCounters", "id", "name")

		obj.processCounters = make(map[int]string)
		for data.Next(&id, &name) {
			obj.processCounters[id] = name
		}

		data.Next()

	}
}
