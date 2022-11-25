package sample

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/leihenshang/http-little-toy/data"
)

// 自定义一个模板类型
type Template string

var template Template = `
{
    "url": "http://localhost:9090/shop/create",
    "method": "POST",
    "header": [
        "Content-Type:application/x-www-form-urlencoded",
        "date:Sun, 09 Oct 2022 03:00:13 GMT"
    ],
    "body": "user_code=127"
}
`

func GenerateRequestFile(fileWithPath string) (err error) {
	if _, statErr := os.Stat(fileWithPath); statErr == nil {
		err = errors.New("the file already exists")
		return
	}

	file, createErr := os.Create(fileWithPath)
	if createErr != nil {
		err = errors.New("the error occurred while creating the file:" + createErr.Error())
		return
	}

	_, err = file.WriteString(string(template))
	return
}

func GenerateRequestFileV1(fileWithPath string, requestSample *data.RequestSample) (err error) {
	if requestSample == nil {
		err = errors.New("request sample object can not be nil")
		return
	}

	if _, statErr := os.Stat(fileWithPath); statErr == nil {
		err = errors.New("the file already exists")
		return
	}

	file, createErr := os.Create(fileWithPath)
	if createErr != nil {
		err = errors.New("the error occurred while creating the file:" + createErr.Error())
		return
	}

	// if requestSample.ExecuteCount <= 0 {
	// 	requestSample.ExecuteCount = 1
	// }
	if requestSample.Request.Url == "" {
		req := &data.Request{}
		unmarshalErr := json.Unmarshal([]byte(template), &req)
		if unmarshalErr != nil {
			return unmarshalErr
		}
		requestSample.Request = *req
	}

	marshalData, marshalErr := json.Marshal(requestSample)
	if marshalErr != nil {
		return marshalErr
	}

	_, err = file.Write(marshalData)
	return
}
