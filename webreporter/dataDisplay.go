package webreporter

import (
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"
)

func (obj *webReporter) dataDisplayPage(w http.ResponseWriter, req *http.Request) {
	obj.counters = obj.listCounters()

	url := req.URL.String()
	// details := obj.getRootDetails()

	data := struct {
		Title, Version                 string
		// ProcessingSize, ProcessingTime string
		// ProcessingSpeed                string
		// FirstEventTime, LastEventTime  string
		DataFilter                     string
		MainMenu                       string
		Series                         map[int]string
	}{
		Title:           obj.title,
		Version:         obj.version,
		// ProcessingSize:  byteCount(details.ProcessingSize),
		// ProcessingTime:  details.ProcessingTime.Format("2006-01-02 15:04:05"),
		// ProcessingSpeed: byteCount(details.ProcessingSpeed),
		// FirstEventTime:  details.FirstEventTime.Format("2006-01-02 15:04:05"),
		// LastEventTime:   details.LastEventTime.Format("2006-01-02 15:04:05"),
		DataFilter:      obj.filter.get(url),
		MainMenu:        obj.mainMenu.get(url),
		//Processes:       toDataRows(obj.getProcesses()),
		Series: obj.counters,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := obj.templates.ExecuteTemplate(w, "dataDisplay.html", data)
	checkErr(err)
}

///////////////////////////////////////////////////////////////////////////////

func (obj *webReporter) listCounters() map[int]string {

	res := make(map[int]string, 0)

	details := obj.storage.Select("counters", "id", "fullName")
	// 	details.SetTimeFilter(obj.filter.getData())
	details.SetFilter("active = TRUE")
	details.SetOrder("id")

	var id int
	var fullName string
	for details.Next(&id, &fullName) {
		res[id] = fullName
	}

	return res
}

///////////////////////////////////////////////////////////////////////////////

// TODO управление маштабом в счётчиках

func (obj *webReporter) getCounterSeries(id string) string {

	rows := make([]string, 0)

	details := obj.storage.Select("dataPoints", "timeStamp", "value")
	details.SetTimeFilter(obj.filter.getData())
	details.SetFilter("counter = ?", id)
	details.SetOrder("timeStamp")

	var timeStamp time.Time
	var value float64
	for details.Next(&timeStamp, &value) {
		rows = append(rows, fmt.Sprintf("[\"%s\", %g]",
			timeStamp.Format("2006-01-02 15:04:05"), value))
	}

	return fmt.Sprintf("{\"id\": %s, \"data\": [%s]}",
		id,
		strings.Join(rows, ","))
}

func (obj *webReporter) getCountersStatistics() string {

	rows := make([]string, 0)

	var counter int
	var cMin, cMax, cAvg, cCount float64

	details := obj.storage.Select("dataPoints", "counter",
		"MIN(value)", "MAX(value)", "AVG(value), COUNT(*)",
	)
	details.SetTimeFilter(obj.filter.getData())
	details.SetFilter("counter IN (SELECT id FROM counters WHERE active = TRUE)")
	details.SetGroup("counter")
	details.SetOrder("counter")

	for details.Next(
		&counter,
		&cMin, &cMax, &cAvg, &cCount,
	) {
		rows = append(rows, fmt.Sprintf(
			"{\"Name\": \"%s\", \"Min\": %g, \"Max\": %g, \"Avg\": %g, \"Count\": %g}",
			template.JSEscapeString(obj.counters[counter]),
			cMin, cMax, cAvg, cCount,
		))
	}

	return fmt.Sprintf("[%s]", strings.Join(rows, ","))
}

func byteCount(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%db", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cb",
		float64(b)/float64(div), "kMGTPE"[exp])
}
