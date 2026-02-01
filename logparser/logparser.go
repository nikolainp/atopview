package logparser

import (
	"context"
	"fmt"
	"time"
)

// LogParser ...
type LogParser interface {
	WithMonitor(Monitor)
	ReadData(context.Context, <-chan []byte) error
}

// Monitor ...
type Monitor interface {
	DiscoveredData(mark string, count int, size int64)
	ProcessedData(mark string, count int, size int64)
}

///////////////////////////////////////////////////////////////////////////////

type logParser struct {
	monitor Monitor
}

// NewLogParser ...
func NewLogParser() LogParser {
	obj := new(logParser)
	return obj
}

///////////////////////////////////////////////////////////////////////////////

func (obj *logParser) WithMonitor(monitor Monitor) {
	obj.monitor = monitor
}

func (obj *logParser) ReadData(ctx context.Context, in <-chan []byte) error {

	isFirstLine := true

	readRecord := func(buf []byte) {
		data := newEntry(buf)

		if isFirstLine {
			isFirstLine = false
			obj.monitor.DiscoveredData(data.timeStamp.String(), 0, 0)
		}
		obj.monitor.ProcessedData(data.timeStamp.String(), 0, 0)

	}

	for isBreak := false; !isBreak; {
		select {
		case <-ctx.Done():
			return nil
		case buf, ok := <-in:
			if ok {
				fmt.Println(string(buf))
				readRecord(buf)
			} else {
				isBreak = true
			}
		}
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////

type entryType int

const (
	cpu entryType = iota
)

// cpu chuwi 1769451659 2026/01/26 21:20:59 432949 100 0 81874 265308 1771 2603194 13107 0 4832 0 0 2399 70 0 0
type dataEntry struct {
	kind      entryType
	computer  string
	timeStamp time.Time
	points    []float64
}

func newEntry(buf []byte) dataEntry {
	return dataEntry{}
}
