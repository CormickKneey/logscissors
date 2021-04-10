package logrotator

import (
	_ "fmt"
	"github.com/lestrrat/go-strftime"
	"github.com/pkg/errors"
	"io"
	"os"
	"path"
	"sync"
	"time"
)

type TimeBasedRotator struct {
	timeDiffToUTC int64
	lastTime      int64
	period        int64
	dirname       string
	filename      string
	mutex         sync.RWMutex
	outFile       *os.File
	pattern       *strftime.Strftime
	preFilename   string
}

func (tw *TimeBasedRotator) Write(p []byte) (n int, err error) {
	tw.mutex.Lock()
	defer tw.mutex.Unlock()

	fh, err := tw.getHandler()
	if err != nil {
		return 0, errors.Wrap(err, `failed to acquire target io.Writer`)
	}

	if fh == nil {
		return 0, errors.Wrap(err, `target io.Writer is closed`)
	}

	return fh.Write(p)
}

func (tw *TimeBasedRotator) getHandler() (io.Writer, error) {
	if tw.preFilename == "" {
		return tw.handler()
	}
	return tw.handlerWithPreFilename()
}

func (tw *TimeBasedRotator) handler() (io.Writer, error) {
	isOvertime, current := tw.isOvertime()
	if !isOvertime {
		return tw.outFile, nil
	}
	filename := tw.pattern.FormatString(time.Unix(0, current))
	//fmt.Printf("FormatString filename=%s\n", filename)
	if tw.filename == filename {
		return tw.outFile, nil
	}
	//fmt.Printf("Rotate filename=%s %s\n", filename, path.Dir(filename))
	dirname := path.Dir(filename)
	if tw.dirname != dirname {
		_ = os.MkdirAll(path.Dir(filename), os.ModePerm)
		tw.dirname = dirname
	}

	fh, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, errors.Errorf("failed to open file %s: %s", tw.pattern, err)
	}

	_ = tw.outFile.Close()
	tw.outFile = fh
	tw.filename = filename
	tw.lastTime = current

	return fh, nil
}

func (tw *TimeBasedRotator) handlerWithPreFilename() (io.Writer, error) {
	isOvertime, current := tw.isOvertime()
	if !isOvertime {
		return tw.outFile, nil
	}
	filename := tw.pattern.FormatString(time.Unix(0, tw.lastPeriod()))
	dirname := path.Dir(filename)
	if tw.dirname != dirname {
		_ = os.MkdirAll(path.Dir(filename), os.ModePerm)
		tw.dirname = dirname
	}

	fh, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, errors.Errorf("failed to open file %s: %s", tw.pattern, err)
	}
	io.Copy(fh, tw.outFile)
	// redo log
	fh.Close()
	tw.outFile.Close()
	os.Truncate(tw.preFilename, 0)

	// reopen
	fl, err := os.OpenFile(tw.preFilename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return nil, errors.Errorf("failed to open file %s: %s", tw.preFilename, err)
	}

	tw.outFile = fl
	tw.filename = filename
	tw.lastTime = current

	return fl, nil
}

func (tw *TimeBasedRotator) Close() error {
	tw.mutex.Lock()
	defer tw.mutex.Unlock()

	if tw.outFile != nil {
		err := tw.outFile.Close()
		if err != nil {
			return err
		}
		tw.outFile = nil
	}
	return nil
}

func (tw *TimeBasedRotator) isOvertime() (bool, int64) {
	nowUnixNano := time.Now().UnixNano()
	current := nowUnixNano - ((nowUnixNano + tw.timeDiffToUTC) % tw.period)
	return (current - tw.lastTime) >= tw.period, current
}

func (tw *TimeBasedRotator) lastPeriod() int64 {
	lastUnixNano := time.Now().Add(-time.Duration(tw.period)).UnixNano()
	return lastUnixNano - ((lastUnixNano + tw.timeDiffToUTC) % tw.period)
}
