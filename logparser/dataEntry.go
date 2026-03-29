package logparser

import (
	"bytes"
	"strconv"
	"time"
)

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

	res.label = getEntryLabel(bufSlice)
	if res.label == labelNONE {
		//err = fmt.Errorf("Unknown label type")
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

	if bytes.ContainsRune(buf, '(') {

		res.points = make([][]byte, 0, len(bufSlice)-6)
		var start int
		var isParenthesis bool
		for i, word := range bufSlice[6:] {
			switch {
			case !isParenthesis && word[0] == '(' && word[len(word)-1] == ')':
				res.points = append(res.points, bytes.Trim(word, "()"))
			case !isParenthesis && word[0] == '(':
				isParenthesis = true
				start = i
			case isParenthesis && word[len(word)-1] == ')':
				isParenthesis = false
				res.points = append(res.points, bytes.Trim(
					bytes.Join(bufSlice[start+6:i+7], []byte(" ")), "()"))
			case !isParenthesis:
				res.points = append(res.points, word)
			}
		}

	} else {
		res.points = bufSlice[6:]
	}

	switch res.label {
	case labelPSI:
		if string(res.points[0]) != "y" {
			res.label = labelNONE
		}
	case labelNET1:
		if string(res.points[0]) != "upper" {
			res.label = labelNET2
		}
	case labelPRG:
		if string(res.points[21]) != "y" {
			res.label = labelNONE
		}
	case labelPRC:
		if string(res.points[13]) != "y" {
			res.label = labelNONE
		}
	}

	return
}

func getEntryLabel(buf [][]byte) entryLabel {
	switch string(buf[0]) {
	case "RESET":
		return labelRESET
	case "SEP":
		return labelSEP
	}

	if len(buf) < 6 {
		return labelNONE
	}

	switch string(buf[0]) {
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
	case "NFC":
		return labelNFC
	case "NFS":
		return labelNFS
	case "NET":
		return labelNET1
	case "NUM":
		return labelNUM
	case "PRG":
		return labelPRG
	case "PRC":
		return labelPRC
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
