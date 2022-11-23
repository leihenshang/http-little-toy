package mylog

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"time"

	fileUtil "github.com/leihenshang/http-little-toy/common/utils/file-util"
)

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

func LogStart(ctx context.Context, logDir string, logChan <-chan []byte) (err error) {

	logFile, logErr := createLog(logDir)
	if logErr != nil {
		return logErr
	}

	// 启动一个协程来处理日志写入
	go func(logCtx context.Context) {
	LOOP:
		for {
			select {
			case l := <-logChan:
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
		// 关闭日志文件
		logFile.Close()
	}(ctx)

	return
}
