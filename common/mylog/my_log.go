package mylog

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

	timeUtil "github.com/leihenshang/http-little-toy/common/utils/time-util"

	fileUtil "github.com/leihenshang/http-little-toy/common/utils/file-util"
)

// MyLog a log object
type MyLog struct {
	logChan chan []byte
	// a counter
	MyWait *sync.WaitGroup
}

// NewMyLog create a `MyLog` object.
func NewMyLog() *MyLog {
	return &MyLog{
		logChan: make(chan []byte),
		MyWait:  &sync.WaitGroup{},
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
func (m *MyLog) Start(ctx context.Context, logDir string) (err error) {
	logFile, logErr := logInit(logDir)
	if logErr != nil {
		return logErr
	}

	go func() {
	LOOP:
		for {
			select {
			case l := <-m.logChan:
				m.MyWait.Done()
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

func (m *MyLog) Write(l []byte) {
	m.logChan <- l
}
