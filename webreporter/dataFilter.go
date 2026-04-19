package webreporter

import (
	"strings"
	"text/template"
	"time"
)

type dataFilter struct {
	htmlTemplate *template.Template

	minimumTime, startTime  time.Time
	maximumTime, finishTime time.Time
}

///////////////////////////////////////////////////////////////////////////////

func newDataFilter(html *template.Template, start, finish time.Time) *dataFilter {
	obj := new(dataFilter)

	obj.htmlTemplate = html
	obj.minimumTime = start
	obj.maximumTime = finish

	return obj
}

///////////////////////////////////////////////////////////////////////////////

func (obj *dataFilter) set(start, finish time.Time) {
	obj.startTime = start
	obj.finishTime = finish
}

func (obj *dataFilter) get(url string) string {
	w := new(strings.Builder)

	data := struct {
		URL                     string
		MinimumTime, StartTime  string
		MaximumTime, FinishTime string
	}{
		URL:         url,
		MinimumTime: obj.minimumTime.Format("2006-01-02T15:04"),
		StartTime:   obj.startTime.Format("2006-01-02T15:04"),
		MaximumTime: obj.maximumTime.Format("2006-01-02T15:04"),
		FinishTime:  obj.finishTime.Format("2006-01-02T15:04"),
	}

	err := obj.htmlTemplate.Execute(w, data)
	checkErr(err)

	return w.String()
}

///////////////////////////////////////////////////////////////////////////////

func (obj *dataFilter) getData() (filter struct{ From, To time.Time }) {
	filter.From = obj.startTime
	filter.To = obj.finishTime

	return
}

// func (obj *dataFilter) getStartTime(tt time.Time) time.Time {
// 	if obj.startTime.Before(tt) {
// 		return tt
// 	}

// 	return obj.startTime
// }

// func (obj *dataFilter) getFinishTime(tt time.Time) time.Time {
// 	if obj.finishTime.After(tt) {
// 		return tt
// 	}

// 	return obj.finishTime
// }
