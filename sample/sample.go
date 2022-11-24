package sample

import (
	"errors"
	"os"
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
		err = errors.New("The file already exists.")
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

func GenerateRequestFileV1(fileWithPath string) (err error) {
	if _, statErr := os.Stat(fileWithPath); statErr == nil {
		err = errors.New("The file already exists.")
		return
	}

	file, createErr := os.Create(fileWithPath)
	if createErr != nil {
		err = errors.New("the error occurred while creating the file:" + createErr.Error())
		return
	}

	sampleData := data.RequestSample{}
	 marshalData, marshalErr := json.Marshal(sampleData)
	 if marshalErr != nil {
		return marshalErr
	}


	_, err = file.WriteBytes(marshalData)
	return
}

