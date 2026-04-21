package webreporter

import (
	"fmt"
	"net/http"
	"strings"
	"text/template"
)

func (obj *webReporter) processCountersPage(w http.ResponseWriter, req *http.Request) {
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
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := obj.templates.ExecuteTemplate(w, "processCounters.html", data)
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
// 	details.SetFilter("active = TRUE")
// 	details.SetOrder("id")

// 	var id int
// 	var fullName string
// 	for details.Next(&id, &fullName) {
// 		res[id] = fullName
// 	}

// 	return res
// }

///////////////////////////////////////////////////////////////////////////////

// TODO сброс всех выделенных
// TODO фильтрация по копмьютеру
// TODO фильтрация по процессу
// TODO фильтрация по типу процесса (кластер, вебсервер, СУБД)
// TODO сортировка групп

func (obj *webReporter) getProcessCountersList() string {

	rows := make([]string, 0)

	details := obj.storage.Select("processCounters", "id", "active", "fullName",
		"label", "name", "description")
	// 	//details.SetTimeFilter(obj.filter.getData())
	// 	details.SetFilter("counter = ?", id)
	details.SetOrder("label")

	var id int64
	var active bool
	var fullName, label, name, description string
	for details.Next(&id, &active, &fullName, &label, &name, &description) {
		rows = append(rows, fmt.Sprintf(
			"{\"ID\": \"%d\", \"Enable\": %t, \"FullName\": \"%s\", \"Label\": \"%s\", \"Name\": \"%s\", \"Description\": \"%s\"}",
			id, active,
			template.JSEscapeString(fullName),
			label, name, description,
		))
	}

	return fmt.Sprintf("[%s]",
		strings.Join(rows, ","))
}

func (obj *webReporter) setProcessCounterActive(id, active string) {

	var argActive bool

	switch active {
	case "true":
		argActive = true
	case "false":
		argActive = false
	}

	{
		update := obj.storage.Update("processCounters", "active", argActive)
		update.SetFilter(fmt.Sprintf("id = %s", id))
		update.Execute()
	}

	{
		update := obj.storage.Update("processCountersData", "active", true)
		update.SetFilter("counter IN (SELECT id FROM processCounters WHERE active = TRUE)")
		update.SetFilter("process IN (SELECT id FROM processInfo WHERE active = TRUE)")
		update.Execute()
	}

}
