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

func createLog(LogDir string) (f *os.File, err error) {
	logDir, logDirErr := fileUtil.IsExisted(LogDir)
	if logDirErr != nil {
		err = errors.New(fmt.Sprintf("an error occurred while get log directory information. err:%+v \n", logDirErr))
		return
	}
	// 日志目录不存在
	if logDir == false {
		dirErr := os.MkdirAll(LogDir, os.ModePerm)
		if dirErr != nil {
			err = errors.New(fmt.Sprintf("an error occurred while make directory.err:%+v \n", dirErr))
			return
		}
	}

	logName := fmt.Sprintf("httpLittleToy-%s.log", time.Now().Format("20060102150405"))
	logPath := path.Join(LogDir, logName)
	logFile, logErr := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE, fs.ModePerm)
	if logErr != nil {
		err = errors.New(fmt.Sprintf("an error occurred while create log file.err:%+v \n", logErr))
		return
	}

	f = logFile
	return
}

// LogStart start log
func (m *MyLog) LogStart(ctx context.Context, logDir string) (err error) {

	logFile, logErr := createLog(logDir)
	if logErr != nil {
		return logErr
	}

	go func(logCtx context.Context) {
	LOOP:
		for {
			select {
			case l := <-m.logChan:
				m.MyWait.Done()
				logData := []byte(time.Now().Format("2006-01-02 15:04:05 "))
				logData = append(logData, l...)
				logData = append(logData, []byte("\n")...)
				_, lErr := logFile.Write(logData)
				if lErr != nil {
					log.Printf("[LogStart] write log err:%+v\n", lErr)
				}
			case <-logCtx.Done():
				break LOOP
			}

		}

		logFile.Close()
	}(ctx)

	return
}

// WriteLog write a log information
func (m *MyLog) WriteLog(l []byte) {
	m.logChan <- l
}
