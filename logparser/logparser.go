package logparser

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"time"
)

// LogParser ...
type LogParser interface {
	WithMonitor(Monitor)
	ReadData(context.Context, <-chan []byte) error
}

// Monitor ...
type Monitor interface {
	WriteEvent(frmt string, args ...any)
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
		data, err := newEntry(buf)
		if err != nil {
			obj.monitor.WriteEvent("Error: %v\n%s\n", err, string(buf))
			return
		}

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
				// fmt.Println(string(buf))
				readRecord(buf)
			} else {
				isBreak = true
			}
		}
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////

type entryLabel int

const (
	labelNONE entryLabel = iota
	labelSEP
	labelRESET
	labelCPUTotal
	labelCPU
	labelCPL
	labelMEM
	labelSWP
	labelPAG
	labelPSI
	labelDSK
	labelNFC
	labelNFS
	labelNET
	labelNUM
	labelPRG
	labelPRC
	labelPRE
	labelPRM
	labelPRD
	labelPRN
)

// cpu chuwi 1769451659 2026/01/26 21:20:59 432949 100 0 81874 265308 1771 2603194 13107 0 4832 0 0 2399 70 0 0
type dataEntry struct {
	label     entryLabel
	computer  string
	timeStamp time.Time
	interval  int64
	points    []float64
}

func newEntry(buf []byte) (res dataEntry, err error) {
	bufSlice := bytes.Split(buf, []byte(" "))

	// 0 - label (the name of the label),
	// 1 - host (the name of this machine),
	// 2 - unix epoch,
	// 3 - date  (date of this interval in format YYYY/MM/DD),
	// 4- time (time of this interval in format HH:MM:SS), and
	// 5 - interval (number of seconds elapsed for this interval).

	res.label = getEntryLabel(bufSlice[0])
	if res.label == labelNONE {
		err = fmt.Errorf("Unknown label type")
		return
	}
	if res.label == labelRESET || res.label == labelSEP {
		return
	}
	res.computer = string(bufSlice[1])
	if res.timeStamp, err = bytesToTime(bufSlice[2]); err != nil {
		return
	}
	if res.interval, err = bytesToInt64(bufSlice[5]); err != nil {
		return
	}

	return
}

func getEntryLabel(buf []byte) entryLabel {
	switch string(buf) {
	case "RESET":
		return labelRESET
	case "SEP":
		return labelSEP
	case "CPU":
		return labelCPUTotal
	case "cpu":
		return labelCPU
	case "CPL":
		return labelCPL
	case "MEM":
		return labelMEM
	case "SWP":
		return labelSWP
	case "PAG":
		return labelPAG
	case "PSI":
		return labelPSI
	case "DSK":
		return labelDSK
	case "PRC":
		return labelPRC
	case "NFC":
		return labelNFC
	case "NFS":
		return labelNFS
	case "NET":
		return labelNET
	case "PRG":
		return labelPRG
	case "NUM":
		return labelNUM
	case "PRE":
		return labelPRE
	case "PRM":
		return labelPRM
	case "PRD":
		return labelPRD
	case "PRN":
		return labelPRN

	}
	return labelNONE
}

///////////////////////////////////////////////////////////////////////////////

func bytesToInt64(b []byte) (int64, error) {
	// Convert byte slice to string
	s := string(b)

	// Parse the string as a base-10 integer with 64-bit size
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}

	return i, nil
}

func bytesToTime(b []byte) (time.Time, error) {
	i, err := bytesToInt64(b)

	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(i, 0), nil
}
