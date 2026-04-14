package logparser

import (
	"context"
	"fmt"
	"sync"
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

	err      error
	transfer *dataTransfer

	desc             map[entryLabel]dataDescription
	details          dataDetails
	systemCounterID  map[string]int
	processCounterID map[string]int
	computerInfo     map[string]*computerInfo
	processInfo      map[struct {
		computer int
		name     string
	}]*processInfo

	processCountersDataID map[struct{ counter, process int }]int
}

// NewLogParser ...
func NewLogParser(storage Storage) LogParser {
	obj := new(logParser)
	obj.storage = storage

	obj.desc = getDataDescription()
	obj.systemCounterID = make(map[string]int, 100)
	obj.processCounterID = make(map[string]int, 100)
	obj.computerInfo = make(map[string]*computerInfo)
	obj.processInfo = make(map[struct {
		computer int
		name     string
	}]*processInfo, 1000)
	obj.processCountersDataID = make(map[struct {
		counter int
		process int
	}]int, 10000)

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

	var wg sync.WaitGroup
	obj.transfer = newDataTransfer(ctx.Done, 10000)

	wg.Go(func() { obj.doRead(ctx, in) })
	wg.Go(func() { obj.doWrite() })

	wg.Wait()

	return obj.err
}

///////////////////////////////////////////////////////////////////////////////

type dataTransfer struct {
	done func() <-chan struct{}
	ch   chan struct {
		table string
		args  []any
	}
}

func newDataTransfer(done func() <-chan struct{}, size int) *dataTransfer {
	obj := new(dataTransfer)
	obj.done = done
	obj.ch = make(chan struct {
		table string
		args  []any
	}, size)
	return obj
}
func (obj *dataTransfer) Close() {
	close(obj.ch)
}
func (obj *dataTransfer) Send(table string, args ...any) (ok bool) {
	select {
	case <-obj.done():
		return false
	case obj.ch <- struct {
		table string
		args  []any
	}{table: table, args: args}:
		return true
	}
}
func (obj *dataTransfer) Receive() (struct {
	table string
	args  []any
}, bool) {
	select {
	case <-obj.done():
		return struct {
			table string
			args  []any
		}{}, false
	case data, ok := <-obj.ch:
		if !ok {
			return struct {
				table string
				args  []any
			}{}, false
		}
		return data, ok
	}
}

///////////////////////////////////////////////////////////////////////////////

func (obj *logParser) doRead(ctx context.Context, in DataTransfer) {

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
				obj.err = err
				return
			}
			in.Free(buf)
		}
	}

	obj.saveDetails()
	obj.saveComputers()
	obj.saveProcesses()

	obj.transfer.Close()

}

func (obj *logParser) doWrite() {
	for {
		row, ok := obj.transfer.Receive()
		if !ok {
			break
		}
		obj.storage.WriteRow(row.table, row.args...)
	}
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

		obj.transfer.Send("dataPoints", data.timeStamp, id, counter.value)
	}

	if data.label == labelPRG {
		notes, err := desc.getNotes(data)
		if err != nil {
			return err
		}

		process := obj.getProcess(computer.getID(), subName)
		process.set(notes)
	}

	// TODO детали загрузки
	// TODO параметры процесса
	// TODO счётчики по всем процессам - число зомби, число процессов этого типа, ждущих процессов,новых/старых/завершившихся

	return nil
}

func (obj *logParser) saveDetails() {

	obj.transfer.Send("dataFilter",
		obj.details.firstEventTime,
		obj.details.lastEventTime,
	)
	obj.transfer.Send("details",
		obj.details.title,
		obj.details.version,
		obj.details.firstEventTime,
		obj.details.lastEventTime,
	)

}

func (obj *logParser) saveComputers() {

	for _, info := range obj.computerInfo {
		obj.transfer.Send("computers",
			info.getID(), info.getName())
	}
}

func (obj *logParser) saveProcesses() {

	for _, info := range obj.processInfo {
		obj.transfer.Send("processInfo",
			info.getID(), true,
			info.computer, info.pid, info.ppid,
			info.name, info.commandLine,
			info.exitCode,
			info.startTime, info.endTime,
		)
	}
}

///////////////////////////////////////////////////////////////////////////////

func (obj *logParser) getComputer(name string) *computerInfo {
	if res, ok := obj.computerInfo[name]; ok {
		return res
	}

	id := len(obj.computerInfo) + 1
	res := newComputerInfo(id, name)
	obj.computerInfo[name] = res

	return res
}

func (obj *logParser) getProcess(computer int, name string) *processInfo {

	key := struct {
		computer int
		name     string
	}{computer, name}

	if res, ok := obj.processInfo[key]; ok {
		return res
	}

	id := len(obj.processInfo) + 1
	res := newProcessInfo(id, computer, name)
	obj.processInfo[key] = res

	return res
}

func (obj *logParser) getCounterID(desc dataDescription, computer *computerInfo, name, subName string) int {

	var id int
	var longName string

	label := desc.getLabel()

	if desc.isSystem {
		if subName == "" {
			longName = fmt.Sprintf("%s^%s^%s",
				computer.getName(), label, name)
		} else {
			longName = fmt.Sprintf("%s^%s^%s^%s",
				computer.getName(), label, name, subName)
		}

		if id, ok := obj.systemCounterID[longName]; ok {
			return id
		}

		id = len(obj.systemCounterID) + len(obj.processCountersDataID) + 1
		details := desc.getDetails(name)

		obj.transfer.Send("computerCounters", id,
			details.active,
			longName, computer.getID(),
			label, name, subName,
			details.description)
		obj.systemCounterID[longName] = id

		if details.isProperty {
			obj.transfer.Send("computerInfo", computer.getID(), id)
		}

	} else {
		longName = fmt.Sprintf("%s^%s^%s",
			computer.getName(), label, name)

		counterID, ok := obj.processCounterID[longName]
		if !ok {
			counterID = len(obj.processCounterID) + 1
			details := desc.getDetails(name)

			obj.transfer.Send("processCounters", counterID,
				details.active,
				longName, computer.getID(),
				label, name,
				details.description)
			obj.processCounterID[longName] = counterID
		}

		process := obj.getProcess(computer.getID(), subName)

		dataKey := struct {
			counter int
			process int
		}{counterID, process.getID()}
		if id, ok := obj.processCountersDataID[dataKey]; ok {
			return id
		}

		id = len(obj.systemCounterID) + len(obj.processCountersDataID) + 1
		obj.processCountersDataID[dataKey] = id

		obj.transfer.Send("processCountersData", counterID, process.getID(), id)
	}

	return id
}

///////////////////////////////////////////////////////////////////////////////
