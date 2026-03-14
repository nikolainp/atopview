package webreporter

import (
	"fmt"
	"net/http"
	"strings"
)

func (obj *webReporter) informationPage(w http.ResponseWriter, req *http.Request) {

	details := obj.getRootDetails()

	data := struct {
		Title, Version string
		DataFilter     string
		MainMenu       string
		Computers      map[int]string
	}{
		Title:   obj.title,
		Version: details.Version,
		//DataFilter:      obj.filter.getContent(req.URL.String()),
		MainMenu:  obj.mainMenu.getMainMenu("/information"),
		Computers: obj.listComputers(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := obj.templates.ExecuteTemplate(w, "information.html", data)
	checkErr(err)
}

///////////////////////////////////////////////////////////////////////////////

func (obj *webReporter) listComputers() map[int]string {

	res := make(map[int]string, 0)

	details := obj.storage.SelectQuery("computers", "id", "name")
	// 	details.SetTimeFilter(obj.filter.getData())
	details.SetOrder("name")

	var id int
	var name string
	for details.Next(&id, &name) {
		res[id] = name
	}

	return res
}

///////////////////////////////////////////////////////////////////////////////

func (obj *webReporter) getInformation(id string) string {

	rows := make([]string, 0)

	details := obj.storage.SelectQuery("computerInfo",
		"label", "name", "subName",
		"min", "max")
	// 	//details.SetTimeFilter(obj.filter.getData())
	details.SetFilter(fmt.Sprintf("computer = %s", id))
	details.SetOrder("label", "subName")

	var label, name, subName string
	var min, max float64
	for details.Next(&label, &name, &subName, &min, &max) {
		rows = append(rows, fmt.Sprintf(
			"{\"Label\": \"%s\", \"Name\": \"%s\", \"SubName\": \"%s\", \"Min\": %g, \"Max\": %g}",
			label, name, subName, min, max,
		))
	}

	return fmt.Sprintf("[%s]",
		strings.Join(rows, ","))
}
