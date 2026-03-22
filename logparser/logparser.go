package logparser

import (
	"context"
	"fmt"
	"time"
)

// LogParser ...
type LogParser interface {
	WithMonitor(Monitor)
	WithDetails(title, version string)
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
	Update(table string, args ...any) interface {
		SetFilter(filter ...string)
		Execute()
	}
	//      SetIdByGroup(table string, column, group string)
	Select(table string, columns ...string) interface {
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

	desc         map[entryLabel]dataDescription
	details      dataDetails
	counterID    map[string]int
	computerInfo map[string]*computerInfo
}

// NewLogParser ...
func NewLogParser(storage Storage) LogParser {
	obj := new(logParser)
	obj.storage = storage

	obj.desc = getDataDescription()
	obj.counterID = make(map[string]int, 100)
	obj.computerInfo = make(map[string]*computerInfo)

	return obj
}

///////////////////////////////////////////////////////////////////////////////

func (obj *logParser) WithMonitor(monitor Monitor) {
	obj.monitor = monitor
}

func (obj *logParser) WithDetails(title, version string) {
	obj.details.title = title
	obj.details.version = version
}

func (obj *logParser) ReadData(ctx context.Context, in DataTransfer) error {

	isFirstLine := true

	readRecord := func(buf []byte, n int) error {
		data, err := newEntry(buf[:n])
		if err != nil {
			obj.monitor.WriteEvent("Error: %v\n%s\n", err, string(buf))
			return nil
		}

		if data.label == labelNONE || data.label == labelRESET || data.label == labelSEP {
			return nil
		}

		if isFirstLine {
			isFirstLine = false
			obj.monitor.WriteEvent("Start time: %s\n", data.timeStamp.String())
			obj.details.firstEventTime = data.timeStamp
			obj.monitor.DiscoveredData(data.timeStamp.String(), 0, 0)
		}
		obj.monitor.ProcessedData(data.timeStamp.String(), 0, 0)
		obj.details.lastEventTime = data.timeStamp

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

	obj.saveDetails()
	obj.saveComputers()

	return nil
}

///////////////////////////////////////////////////////////////////////////////

func (obj *logParser) saveRecord(data dataEntry) error {

	desc, ok := obj.desc[data.label]
	if !ok {
		return fmt.Errorf("label not found")
	}

	computer := obj.getComputer(data.computer)
	subName := desc.getSubName(data)

	counters, err := desc.getCounters(data)
	if err != nil {
		return err
	}

	for _, counter := range counters {

		id := obj.getCounterID(desc, computer, counter.key, subName)

		obj.storage.WriteRow("dataPoints", data.timeStamp, id, counter.value)
	}

	// TODO детали загрузки
	// TODO параметры процесса
	// TODO счётчики по всем процессам - число зомби, число процессов этого типа, ждущих процессов,новых/старых/завершившихся

	return nil
}

func (obj *logParser) saveDetails() {

	obj.storage.WriteRow("dataFilter",
		obj.details.firstEventTime,
		obj.details.lastEventTime,
	)
	obj.storage.WriteRow("details",
		obj.details.title,
		obj.details.version,
		obj.details.firstEventTime,
		obj.details.lastEventTime,
	)

}

func (obj *logParser) saveComputers() {

	for _, info := range obj.computerInfo {
		obj.storage.WriteRow("computers",
			info.getID(), info.getName())
	}

}

func (obj *logParser) getComputer(name string) *computerInfo {
	if res, ok := obj.computerInfo[name]; ok {
		return res
	}

	id := len(obj.computerInfo) + 1
	res := newComputerInfo(id, name)
	obj.computerInfo[name] = res

	return res
}

func (obj *logParser) getCounterID(desc dataDescription, computer *computerInfo, name, subName string) int {

	var longName string

	label := desc.getLabel()
	if subName == "" {
		longName = fmt.Sprintf("%s^%s^%s",
			computer.getName(), label, name)

	} else {
		longName = fmt.Sprintf("%s^%s^%s^%s",
			computer.getName(), label, name, subName)
	}

	if id, ok := obj.counterID[longName]; ok {
		return id
	}

	id := len(obj.counterID) + 1
	details := desc.getDetails(name)

	obj.storage.WriteRow("counters", id,
		desc.isSystem, details.active,
		longName, computer.getID(),
		label, name, subName,
		details.description)
	obj.counterID[longName] = id

	if details.isProperty {
		obj.storage.WriteRow("computerInfo", computer.getID(), id)
	}

	return id
}

///////////////////////////////////////////////////////////////////////////////
