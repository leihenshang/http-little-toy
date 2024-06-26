package file

import "os"

func IsFileExisted(name string) (existed bool) {
	if _, err := os.Stat(name); err != nil {
		if os.IsExist(err) {
			return true
		}

		if os.IsNotExist(err) {
			return false
		}
	}

	return false
}
