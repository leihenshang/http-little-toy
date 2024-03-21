package mylog

import (
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
	if fileUtil.IsFileExisted(LogDir) == false {
		if dirErr := os.MkdirAll(LogDir, os.ModePerm); dirErr != nil {
			err = errors.New(fmt.Sprintf("an error occurred while make directory.err:%+v \n", dirErr))
			return
		}
	}

	logName := fmt.Sprintf("httpLittleToy-%s.log", time.Now().Format(timeUtil.DateTimeFormat))
	logPath := path.Join(LogDir, logName)
	logFile, logErr := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE, fs.ModePerm)
	if logErr != nil {
		err = errors.New(fmt.Sprintf("an error occurred while create log file.err:%+v \n", logErr))
		return
	}

	f = logFile
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
				logData := []byte(time.Now().Format(timeUtil.DateTimeFormat))
				logData = append(logData, l...)
				logData = append(logData, []byte("\n")...)
				_, lErr := logFile.Write(logData)
				if lErr != nil {
					log.Printf("[Start] write log err:%+v\n", lErr)
				}
			case <-ctx.Done():
				break LOOP
			}

		}

		logFile.Close()
	}()

	return
}

// WriteLog write a log information
func (m *MyLog) WriteLog(l []byte) {
	m.logChan <- l
}
