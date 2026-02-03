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
	ReadData(context.Context, chan<- []byte) (string, error)
}

///////////////////////////////////////////////////////////////////////////////

type logReader struct {
	pathUtility string
	pathLog     string

	err    error
	stderr io.ReadCloser

	bufSize int
	command *exec.Cmd
	poolBuf sync.Pool
	chBuf   chan streamBuffer
}

// NewLogReader ...
func NewLogReader(utility, log string) LogReader {
	obj := new(logReader)

	obj.pathUtility = utility
	obj.pathLog = log

	obj.bufSize = 1024 * 1024 * 1

	obj.poolBuf = sync.Pool{New: func() interface{} {
		lines := make([]byte, obj.bufSize)
		return &lines
	}}

	return obj
}

///////////////////////////////////////////////////////////////////////////////

func (obj *logReader) ReadData(ctx context.Context, out chan<- []byte) (errText string, err error) {
	var data io.Reader

	if data, err = obj.openLogFile(); err != nil {
		return "", err
	}

	var wg sync.WaitGroup
	obj.chBuf = make(chan streamBuffer, 1)

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
		buf := obj.poolBuf.Get().(*[]byte)
		n, err := readBuffer(*buf)
		if n == 0 || err != nil {
			obj.err = err
			break
		} else {
			select {
			case obj.chBuf <- streamBuffer{buf, n}:
				break
			case <-ctx.Done():
				isBreak = true
			}
		}
	}

	close(obj.chBuf)
}

func (obj *logReader) doWrite(ctx context.Context, out chan<- []byte) {

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
		select {
		case buffer, ok := <-obj.chBuf:
			if ok {
				writeBuffer(*(buffer.buf), buffer.len)

				obj.poolBuf.Put(buffer.buf)
			} else {
				if isExistsLastLine {
					writeToChan(out, lastLine)
				}
				isBreak = true
			}
		case <-ctx.Done():

			isBreak = true
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

func writeToChan(out chan<- []byte, buf []byte) {

	bufCopy := make([]byte, len(buf))
	copy(bufCopy, buf)
	out <- bufCopy
}
