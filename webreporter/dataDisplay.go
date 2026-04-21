package webreporter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (obj *webReporter) dataDisplayPage(w http.ResponseWriter, req *http.Request) {
	obj.counters = obj.listCounters()

	url := req.URL.String()

	data := struct {
		Title, Version string
		DataFilter     string
		MainMenu       string
		Series         map[int]string
	}{
		Title:      obj.details.Title,
		Version:    obj.details.Version,
		DataFilter: obj.filter.get(url),
		MainMenu:   obj.mainMenu.get(url),
		Series:     obj.counters,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := obj.templates.ExecuteTemplate(w, "dataDisplay.html", data)
	checkErr(err)
}

///////////////////////////////////////////////////////////////////////////////

func (obj *webReporter) listCounters() map[int]string {

	res := make(map[int]string, 0)

	{
		details := obj.storage.Select("systemCounters", "id", "fullName")
		details.SetFilter("active = TRUE")
		details.SetOrder("id")

		var id int
		var fullName string
		for details.Next(&id, &fullName) {
			res[id] = fullName
		}
	}

	{
		details := obj.storage.Select("processCountersData", "data", "fullName")
		details.SetFilter("counter IN (SELECT id FROM processCounters WHERE active = TRUE)")
		details.SetFilter("process IN (SELECT id FROM processInfo WHERE active = TRUE)")
		details.SetOrder("data")

		var id int
		var fullName string
		for details.Next(&id, &fullName) {
			res[id] = fullName
		}
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
	details.SetFilter("counter IN (SELECT id FROM systemCounters WHERE active = TRUE UNION ALL SELECT data FROM processCountersData WHERE active = TRUE)")
	details.SetGroup("counter")
	details.SetOrder("counter")

	for details.Next(
		&counter,
		&cMin, &cMax, &cAvg, &cCount,
	) {
		rows = append(rows, fmt.Sprintf(
			"{\"Name\": \"%s\", \"Min\": %g, \"Max\": %g, \"Avg\": %g, \"Count\": %g}",
			jsonEscape(obj.counters[counter]),
			cMin, cMax, cAvg, cCount,
		))
	}

	return fmt.Sprintf("[%s]", strings.Join(rows, ","))
}

///////////////////////////////////////////////////////////////////////////////

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

func jsonEscape(i string) string {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	// Trim the beginning and trailing " character
	return string(b[1 : len(b)-1])
}
