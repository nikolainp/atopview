package monitor

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

// Monitor ...
type Monitor interface {
	Start(context.Context, string)
	Stop()
	WriteEvent(frmt string, args ...any)
	DiscoveredData(mark string, count int, size int64)
	ProcessedData(mark string, count int, size int64)
}

///////////////////////////////////////////////////////////////////////////////

type monitor struct {
	markStart, markDone   string
	partsTotal, partsDone int
	sizeTotal, sizeDone   int64
	messageBuffer         []string

	fmtShowProgress string

	startTime time.Time
	ticker    *time.Ticker
	done      chan struct{}

	mu  sync.Mutex
	wg  sync.WaitGroup
	ctx context.Context
}

// NewMonitor ...
func NewMonitor() Monitor {
	obj := new(monitor)

	return obj
}

///////////////////////////////////////////////////////////////////////////////

// Start ...
func (obj *monitor) Start(ctx context.Context, fmtShowProgress string) {
	obj.Stop()

	obj.ctx = ctx
	obj.done = make(chan struct{})

	obj.markStart, obj.markDone = "", ""
	obj.partsTotal, obj.partsDone = 0, 0
	obj.sizeTotal, obj.sizeDone = 0, 0
	obj.fmtShowProgress = fmtShowProgress + "          \r"

	obj.print()
}

// Stop ...
func (obj *monitor) Stop() {
	if obj.ticker == nil {
		return
	}

	defer obj.ticker.Stop()

	close(obj.done)
	obj.wg.Wait()
}

func (obj *monitor) WriteEvent(frmt string, args ...any) {
	defer obj.mu.Unlock()
	obj.mu.Lock()

	obj.messageBuffer = append(obj.messageBuffer, fmt.Sprintf(frmt, args...))
}

func (obj *monitor) DiscoveredData(mark string, count int, size int64) {
	defer obj.mu.Unlock()
	obj.mu.Lock()

	obj.markStart = mark
	obj.partsTotal += count
	obj.sizeTotal += size
}

func (obj *monitor) ProcessedData(mark string, count int, size int64) {
	defer obj.mu.Unlock()
	obj.mu.Lock()

	obj.markDone = mark
	obj.partsDone += count
	obj.sizeDone += size
}

///////////////////////////////////////////////////////////////////////////////

func (obj *monitor) print() {
	var prevFinishedSize int64
	var prevDuration time.Duration

	obj.startTime = time.Now()
	obj.ticker = time.NewTicker(500 * time.Millisecond)

	doPrint := func() {
		defer obj.mu.Unlock()
		obj.mu.Lock()

		var speed int64
		var totalSpeed int64

		totalDuration := time.Since(obj.startTime)
		totalSpeed = 1000 * obj.sizeDone / totalDuration.Milliseconds()

		deltaDuration := totalDuration - prevDuration
		if deltaDuration.Milliseconds() > 0 {
			speed = 1000 * (obj.sizeDone - prevFinishedSize) / deltaDuration.Milliseconds()
			if deltaDuration.Seconds() < 1 {
				speed = 1000 * speed / deltaDuration.Milliseconds()
			}
		}
		if deltaDuration.Seconds() > 1 {
			prevDuration = totalDuration
			prevFinishedSize = obj.sizeDone
		}

		for i := range obj.messageBuffer {
			fmt.Fprint(os.Stderr, obj.messageBuffer[i])
		}
		obj.messageBuffer = obj.messageBuffer[:0]

		//"files: %d/%d size: %s/%s time: %s [speed %s/s/%s/s ]                           \r",
		fmt.Fprintf(os.Stderr,
			obj.fmtShowProgress,
			obj.partsDone, obj.partsTotal,
			byteCount(obj.sizeDone), byteCount(obj.sizeTotal),
			obj.markDone, obj.markStart,
			totalDuration.Truncate(time.Second),
			byteCount(speed), byteCount(totalSpeed))
	}

	obj.wg.Go(func() {
		for isBreak := false; !isBreak; {
			select {
			case <-obj.done:
				isBreak = true
			case <-obj.ctx.Done():
				isBreak = true

			case <-obj.ticker.C:
				doPrint()
			}
		}

		doPrint()
		fmt.Fprintf(os.Stderr, "\n")
	})

	// TODO: + total bytesnv time spend [ speed ] [ in work %d - finished %d ]
}

///////////////////////////////////////////////////////////////////////////////

func byteCount(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%db", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cb",
		float64(b)/float64(div), "kMGTPE"[exp])
}
