package toylog

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"sync"
	"time"

	timeUtil "github.com/leihenshang/http-little-toy/common/utils/datetime"
	fileUtil "github.com/leihenshang/http-little-toy/common/utils/file"
)

// ToyLog a log object
type ToyLog struct {
	c chan []byte
	// a counter
	Wait *sync.WaitGroup
}

// NewMyLog create a `ToyLog` object.
func NewMyLog() *ToyLog {
	return &ToyLog{
		c:    make(chan []byte),
		Wait: &sync.WaitGroup{},
	}
}

func logInit(LogDir string) (f *os.File, err error) {
	if !fileUtil.IsFileExisted(LogDir) {
		if err = os.MkdirAll(LogDir, os.ModePerm); err != nil {
			err = errors.New(fmt.Sprintf("an error occurred while make directory.err:%v \n", err))
			return
		}
	}

	logName := fmt.Sprintf("httpLittleToy-%s.log", time.Now().Format(timeUtil.DateTimeFormat))
	logPath := path.Join(LogDir, logName)
	if f, err = os.OpenFile(logPath, os.O_RDWR|os.O_CREATE, fs.ModePerm); err != nil {
		err = errors.New(fmt.Sprintf("an error occurred while create log file.err:%v \n", err))
	}

	return
}

// Start logging
func (m *ToyLog) Start(ctx context.Context, logDir string) (err error) {
	logFile, logErr := logInit(logDir)
	if logErr != nil {
		return logErr
	}

	go func() {
	LOOP:
		for {
			select {
			case l := <-m.c:
				m.Wait.Done()
				var buf bytes.Buffer
				buf.WriteString(time.Now().Format(timeUtil.DateTimeFormat))
				buf.Write(l)
				buf.WriteString("\n")
				if _, err = logFile.Write(buf.Bytes()); err != nil {
					log.Printf("[Start] write log err:%v\n", err)
				}
			case <-ctx.Done():
				break LOOP
			}

		}

		logFile.Close()
	}()

	return
}

func (m *ToyLog) Write(l []byte) {
	m.c <- l
}
