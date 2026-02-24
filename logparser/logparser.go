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
	ReadData(context.Context, DataTransfer) error
}

// DataTransfer ...
type DataTransfer interface {
	Receive(context.Context) (b *[]byte, n int, ok bool)
	Free(b *[]byte)
}

// Storage ...
type Storage interface {
	WriteRow(table string, args ...any)
	Update(table string, args ...any)
	//      SetIdByGroup(table string, column, group string)
	SelectQuery(table string, columns ...string) interface {
		SetTimeFilter(struct {
			From time.Time
			To   time.Time
		})
		SetFilter(filter ...string)
		SetGroup(fields ...string)
		SetOrder(fields ...string)
		Next(args ...any) bool
	}
}

// Monitor ...
type Monitor interface {
	WriteEvent(frmt string, args ...any)
	DiscoveredData(mark string, count int, size int64)
	ProcessedData(mark string, count int, size int64)
}

///////////////////////////////////////////////////////////////////////////////

type logParser struct {
	storage Storage
	monitor Monitor

	desc       map[entryLabel]dataDescription
	counterID  map[string]int
	computerID int
}

// NewLogParser ...
func NewLogParser(storage Storage) LogParser {
	obj := new(logParser)
	obj.storage = storage

	obj.desc = getDataDescription()
	obj.counterID = make(map[string]int, 100)

	return obj
}

///////////////////////////////////////////////////////////////////////////////

func (obj *logParser) WithMonitor(monitor Monitor) {
	obj.monitor = monitor
}

func (obj *logParser) ReadData(ctx context.Context, in DataTransfer) error {

	isFirstLine := true

	readRecord := func(buf []byte, n int) error {
		data, err := newEntry(buf[:n])
		if err != nil {
			obj.monitor.WriteEvent("Error: %v\n%s\n", err, string(buf))
			return err
		}

		if data.label == labelNONE || data.label == labelRESET || data.label == labelSEP {
			return nil
		}

		if isFirstLine {
			isFirstLine = false
			obj.monitor.WriteEvent("Start time: %s\n", data.timeStamp.String())
			obj.monitor.DiscoveredData(data.timeStamp.String(), 0, 0)
		}
		obj.monitor.ProcessedData(data.timeStamp.String(), 0, 0)

		err = obj.saveRecord(data)

		return err
	}

	for isBreak := false; !isBreak; {
		buf, n, ok := in.Receive(ctx)
		if !ok || n == 0 {
			isBreak = true
		} else {
			// fmt.Println(string(buf))
			if err := readRecord(*buf, n); err != nil {
				return err
			}
			in.Free(buf)
		}
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////

func (obj *logParser) saveRecord(data dataEntry) error {

	desc, ok := obj.desc[data.label]
	if !ok {
		return fmt.Errorf("label not found")
	}

	counters, err := desc.getCounters(data)
	if err != nil {
		return err
	}
	subName := desc.getSubName(data)
	for _, counter := range counters {

		id := obj.getCounterID(desc, counter.key, subName)

		obj.storage.WriteRow("dataPoints", data.timeStamp, id, counter.value)
	}

	return nil
}

func (obj *logParser) getCounterID(desc dataDescription, name, subName string) int {

	label := desc.getLabel()
	longName := fmt.Sprintf("%s^%s^%s", label, name, subName)
	if id, ok := obj.counterID[longName]; ok {
		return id
	}

	id := len(obj.counterID) + 1
	details := desc.getDetails(name)

	obj.storage.WriteRow("counters", id, details.enable, longName, obj.computerID, label, name, subName, details.description)
	obj.counterID[longName] = id

	return id
}

///////////////////////////////////////////////////////////////////////////////

// cpu chuwi 1769451659 2026/01/26 21:20:59 432949 100 0 81874 265308 1771 2603194 13107 0 4832 0 0 2399 70 0 0
type dataEntry struct {
	label     entryLabel
	computer  string
	timeStamp time.Time
	interval  int64
	points    [][]byte
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
	res.points = bufSlice[6:]

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

func bytesToFloat64(b []byte) (float64, error) {
	// Convert byte slice to string
	s := string(b)

	// Parse the string as a base-10 integer with 64-bit size
	i, err := strconv.ParseFloat(s, 10)
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
