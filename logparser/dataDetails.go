package logparser

import "time"

type dataDetails struct {
	title, version                string
	firstEventTime, lastEventTime time.Time
}

