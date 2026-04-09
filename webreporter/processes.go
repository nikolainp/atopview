package webreporter

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (obj *webReporter) processesPage(w http.ResponseWriter, req *http.Request) {
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
	err := obj.templates.ExecuteTemplate(w, "processes.html", data)
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

func (obj *webReporter) getProcessesList() string {

	var id, pid, ppid int64
	var active bool
	var name, commandLine, exitCode string
	var startTime, endTime time.Time

	rows := make([]string, 0)
	values := make([]string, 0, len(obj.computerCounters))
	columns := []string{
		"id", "active", "pid", "ppid",
		"name", "commandLine", "exitCode",
		"startTime", "endTime",
	}
	pointers := []interface{}{
		&id, &active, &pid, &ppid,
		&name, &commandLine, &exitCode,
		&startTime, &endTime,
	}
	startColumn := len(columns)

	for i := range obj.processCounters {
		name := obj.processCounters[i]
		columns = append(columns, name+"_min", name+"_max")
		pointers = append(pointers, new(float64), new(float64))
	}

	data := obj.storage.Select("processInfo", columns...)
	// , )
	data.SetTimeFilter(obj.filter.getData())
	// 	details.SetFilter("counter = ?", id)
	data.SetOrder("pid")

	for data.Next(pointers...) {

		for i := startColumn; i < len(pointers); i++ {
			values = append(values, fmt.Sprintf(
				"\"%s\": %g",
				columns[i], *(pointers[12]).(*float64),
			))
		}

		rows = append(rows, fmt.Sprintf(
			"{\"ID\": \"%d\", \"Enable\": %t, \"PID\": %d, \"PPID\": %d, \"Name\": \"%s\", \"CommandLine\": \"%s\", \"ExitCode\": \"%s\", \"StartTime\": \"%s\", \"EndTime\": \"%s\", %s}",
			id, active, pid, ppid,
			name, commandLine, exitCode,
			startTime, endTime,
			strings.Join(values, ", "),
		))

		values = values[:0]
	}

	return fmt.Sprintf("[%s]",
		strings.Join(rows, ","))
}

func (obj *webReporter) setProcessActive(id, active string) {

	var argActive bool

	switch active {
	case "true":
		argActive = true
	case "false":
		argActive = false
	}

	update := obj.storage.Update("processInfo", "active", argActive)
	update.SetFilter(fmt.Sprintf("id = %s", id))
	update.Execute()

}
