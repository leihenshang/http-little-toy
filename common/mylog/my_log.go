package mylog

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"time"

	fileUtil "github.com/leihenshang/http-little-toy/common/utils/file-util"
)

func CreateLog(LogDir string) (f *os.File, err error) {
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
