package logreader

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"
	"sync"
)

// LogReader ...
type LogReader interface {
	WithMonitor(Monitor)
	ReadData(context.Context, DataTransfer) (string, error)
}

// DataTransfer ...
type DataTransfer interface {
	Close()
	GetBuffer() *[]byte
	Send(b *[]byte, n int)
	Receive(context.Context) (b *[]byte, n int, ok bool)
	Free(b *[]byte)
}

// Monitor ...
type Monitor interface {
	WriteEvent(frmt string, args ...any)
}

///////////////////////////////////////////////////////////////////////////////

type logReader struct {
	pathUtility string
	pathLog     string

	err    error
	stderr io.ReadCloser

	bufSize  int
	command  *exec.Cmd
	transfer DataTransfer
	monitor  Monitor
}

// NewLogReader ...
func NewLogReader(utility, log string) LogReader {
	obj := new(logReader)

	obj.pathUtility = utility
	obj.pathLog = log

	obj.bufSize = 1024 * 1024 * 1

	return obj
}

///////////////////////////////////////////////////////////////////////////////

func (obj *logReader) WithMonitor(monitor Monitor) {
	obj.monitor = monitor
}

func (obj *logReader) ReadData(ctx context.Context, out DataTransfer) (errText string, err error) {
	var data io.Reader

	if data, err = obj.openLogFile(); err != nil {
		return "", err
	}

	var wg sync.WaitGroup
	obj.transfer = NewDataTransfer(obj.bufSize)

	wg.Go(func() { obj.doRead(ctx, data) })
	wg.Go(func() { obj.doWrite(ctx, out) })

	wg.Wait()

	if errText, err := obj.closeLogFile(); err != nil {
		return errText, err
	}

	return errText, obj.err
}

///////////////////////////////////////////////////////////////////////////////

func (obj *logReader) openLogFile() (data io.Reader, err error) {
	obj.monitor.WriteEvent("Exec: %s -PALL -r %s\n", obj.pathUtility, obj.pathLog)
	obj.command = exec.Command(obj.pathUtility, "-PALL", "-r", obj.pathLog)

	if data, err = obj.command.StdoutPipe(); err != nil {
		return
	}
	if obj.stderr, err = obj.command.StderrPipe(); err != nil {
		return
	}

	if err = obj.command.Start(); err != nil {
		return
	}
	return
}

func (obj *logReader) closeLogFile() (errText string, err error) {
	if errText, err = readAllStream(obj.stderr); err != nil {
		return
	}

	err = obj.command.Wait()

	if err == nil {
		err = obj.err
	}

	return errText, err
}

///////////////////////////////////////////////////////////////////////////////

type streamBuffer struct {
	buf *[]byte
	len int
}

type dataTransfer struct {
	ch chan struct {
		buf *[]byte
		len int
	}
	pool sync.Pool
}

// NewDataTransfer ...
func NewDataTransfer(size int) DataTransfer {
	obj := new(dataTransfer)
	obj.ch = make(chan struct {
		buf *[]byte
		len int
	})
	obj.pool = sync.Pool{New: func() interface{} {
		buf := make([]byte, size)
		return &buf
	}}
	return obj
}
func (obj *dataTransfer) Close() {
	close(obj.ch)
}

func (obj *dataTransfer) GetBuffer() *[]byte {
	return obj.pool.Get().(*[]byte)
}
func (obj *dataTransfer) Send(b *[]byte, n int) {
	obj.ch <- streamBuffer{b, n}
}
func (obj *dataTransfer) Receive(ctx context.Context) (b *[]byte, n int, ok bool) {
	select {
	case <-ctx.Done():
		return nil, 0, false
	case data, ok := <-obj.ch:
		if !ok {
			return nil, 0, false
		}
		return data.buf, data.len, ok
	}
}
func (obj *dataTransfer) Free(b *[]byte) {
	obj.pool.Put(b)
}

///////////////////////////////////////////////////////////////////////////////

func (obj *logReader) doRead(ctx context.Context, sIn io.Reader) {

	reader := bufio.NewReaderSize(sIn, obj.bufSize)

	readBuffer := func(buf []byte) (int, error) {
		n, err := reader.Read(buf)
		if n == 0 && err == io.EOF {
			return 0, nil
		}
		if err != nil {
			return n, err
		}

		//		fmt.Println(string(buf[:n]))

		return n, nil
	}

	for isBreak := false; !isBreak; {
		buf := obj.transfer.GetBuffer()
		n, err := readBuffer(*buf)
		if n == 0 || err != nil {
			obj.err = err
			break
		} else {

			select {
			case <-ctx.Done():
				isBreak = true
			default:
				obj.transfer.Send(buf, n)
			}
		}
	}

	obj.transfer.Close()
}

func (obj *logReader) doWrite(ctx context.Context, out DataTransfer) {

	lastLine := make([]byte, obj.bufSize*2)
	isExistsLastLine := false

	writeBuffer := func(buf []byte, n int) {
		isLastStringFull := bytes.Equal(buf[n-1:n], []byte("\n"))

		bufSlice := bytes.Split(buf[:n], []byte("\n"))

		for i := range bufSlice {
			select {
			case <-ctx.Done():
				return
			default:

				if i == 0 && isExistsLastLine {
					lastLine = append(lastLine, bufSlice[i]...)
					if len(bufSlice) > 1 {
						writeToChan(out, lastLine)
						isExistsLastLine = false
					}
					continue
				}
				if i == len(bufSlice)-1 {
					if !isLastStringFull {
						lastLine = lastLine[0:len(bufSlice[i])]
						nc := copy(lastLine, bufSlice[i])
						if nc != len(bufSlice[i]) {
							panic(0)
						}
						isExistsLastLine = true
					}
					continue
				}

				writeToChan(out, bufSlice[i])
			}
		}
	}

	for isBreak := false; !isBreak; {
		buf, n, ok := obj.transfer.Receive(ctx)
		if !ok || n == 0 {
			if isExistsLastLine {
				writeToChan(out, lastLine)
			}
			isBreak = true
		} else {
			writeBuffer(*(buf), n)
			obj.transfer.Free(buf)
		}
	}

}

///////////////////////////////////////////////////////////////////////////////

func readAllStream(data io.Reader) (string, error) {
	buf := make([]byte, 1024)
	totalBuf := &bytes.Buffer{}

	for {
		n, err := data.Read(buf)
		if n == 0 && err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		totalBuf.Write(buf[:n])
	}

	return totalBuf.String(), nil
}

func writeToChan(out DataTransfer, buf []byte) {

	bufCopy := out.GetBuffer()
	if len(*bufCopy) < len(buf) {
		b := make([]byte, len(buf))
		bufCopy = &b
	}
	copy(*bufCopy, buf)

	out.Send(bufCopy, len(buf))
}
