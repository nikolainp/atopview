package logparser

import "time"

type processInfo struct {
	id       int
	computer int
	nameID   string

	pid, ppid          int64
	name, commandLine  string
	exitCode           string
	startTime, endTime time.Time
	elapsedTime        int64
}

func newProcessInfo(id, computer int, name string) *processInfo {
	obj := new(processInfo)

	obj.id = id
	obj.computer = computer
	obj.nameID = name

	return obj
}

///////////////////////////////////////////////////////////////////////////////

func (obj *processInfo) getID() int {
	return obj.id
}

func (obj *processInfo) set(data []keyNote) (err error) {

	if obj.name == "" {
		for _, d := range data {
			switch d.key {
			case "pid":
				obj.pid, err = bytesToInt64(d.note)
			case "ppid":
				obj.ppid, err = bytesToInt64(d.note)
			case "name":
				obj.name = string(d.note)
			case "commandLine":
				obj.commandLine = string(d.note)
			case "startTime":
				obj.startTime, err = bytesToTime(d.note)
			}
		}

		if err != nil {
			return err
		}
	}
	for _, d := range data {
		switch d.key {
		case "endTime":
			var endTime time.Time
			endTime, err = bytesToTime(d.note)
			if endTime.After(obj.endTime) {
				obj.endTime = endTime
			}
		case "exitCode":
			obj.exitCode = string(d.note)
		case "elapsedTime":
			obj.elapsedTime, err = bytesToInt64(d.note)
		}

		if err != nil {
			return err
		}
	}

	return
}
