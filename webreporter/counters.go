package webreporter

import (
	"fmt"
	"net/http"
	"strings"
	"text/template"
)

func (obj *webReporter) countersPage(w http.ResponseWriter, req *http.Request) {

	obj.counters = obj.listCounters()
	details := obj.getRootDetails()

	data := struct {
		Title, Version                 string
		ProcessingSize, ProcessingTime string
		ProcessingSpeed                string
		FirstEventTime, LastEventTime  string
		DataFilter                     string
		MainMenu                       string
		Series                         map[int]string
	}{
		Title:   obj.title,
		Version: details.Version,
		//DataFilter:      obj.filter.getContent(req.URL.String()),
		MainMenu: obj.mainMenu.getMainMenu("/counters"),
		//Processes:       toDataRows(obj.getProcesses()),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := obj.templates.ExecuteTemplate(w, "countersPage.html", data)
	checkErr(err)
}

///////////////////////////////////////////////////////////////////////////////

// type rootDetails struct {
// 	Title, Version                                string
// 	ProcessingSize, ProcessingSpeed               int64
// 	ProcessingTime, FirstEventTime, LastEventTime time.Time
// }

// func (obj *webReporter) getRootDetails() (data rootDetails) {

// 	details := obj.storage.SelectQuery("details")
// 	details.Next(
// 		&data.Title, &data.Version,
// 		&data.ProcessingSize, &data.ProcessingSpeed,
// 		&data.ProcessingTime,
// 		&data.FirstEventTime, &data.LastEventTime)

// 	details.Next()

// 	return
// }

// func (obj *webReporter) listCounters() map[int]string {

// 	res := make(map[int]string, 0)

// 	details := obj.storage.SelectQuery("counters", "id", "fullName")
// 	// 	details.SetTimeFilter(obj.filter.getData())
// 	details.SetFilter("enable = TRUE")
// 	details.SetOrder("id")

// 	var id int
// 	var fullName string
// 	for details.Next(&id, &fullName) {
// 		res[id] = fullName
// 	}

// 	return res
// }

///////////////////////////////////////////////////////////////////////////////

func (obj *webReporter) getCountersList() string {

	rows := make([]string, 0)

	details := obj.storage.SelectQuery("counters", "id", "enable", "fullName",
		"label", "name", "subName", "description")
	// 	//details.SetTimeFilter(obj.filter.getData())
	// 	details.SetFilter("counter = ?", id)
	details.SetOrder("label")

	var id int64
	var enable bool
	var fullName, label, name, subName, description string
	for details.Next(&id, &enable, &fullName, &label, &name, &subName, &description) {
		rows = append(rows, fmt.Sprintf(
			"{\"ID\": \"%d\", \"Enable\": %t, \"FullName\": \"%s\", \"Label\": \"%s\", \"Name\": \"%s\", \"SubName\": \"%s\", \"Description\": \"%s\"}",
			id, enable,
			template.JSEscapeString(fullName),
			label, name, subName, description,
		))
	}

	return fmt.Sprintf("[%s]",
		strings.Join(rows, ","))
}

func (obj *webReporter) setCounterEnable(id, enable string) {

	update := obj.storage.Update("counters", "enable", enable)
	update.SetFilter("id = ?", id)
	update.Execute()

}
