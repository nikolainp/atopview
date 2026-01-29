package logparser

import (
	"context"
	"fmt"
)

// LogParser ...
type LogParser interface {
	ReadData(context.Context, <-chan []byte) error
}

///////////////////////////////////////////////////////////////////////////////

type logParser struct {
}

// NewLogParser ...
func NewLogParser() LogParser {
	obj := new(logParser)
	return obj
}

///////////////////////////////////////////////////////////////////////////////

func (obj *logParser) ReadData(ctx context.Context, in <-chan []byte) error {

	for isBreak := false; !isBreak; {
		select {
		case <-ctx.Done():
			return nil
		case buf, ok := <-in:
			if ok {
				fmt.Println(string(buf))
			} else {
				isBreak = true
			}

		}
	}

	return nil
}
