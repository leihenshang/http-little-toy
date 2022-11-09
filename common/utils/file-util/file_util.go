package file_util

import "os"

func IsExisted(path string) (existed bool, err error) {
	_, err = os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true, nil
		}

		if os.IsNotExist(err) {
			return false, nil
		}

	}

	return false, err
}
