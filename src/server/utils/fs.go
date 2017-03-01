package utils

import "os"

func EnsureDirectory(path string) error {
	var err error

	if _, err = os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, os.ModePerm)

		return err
	}

	return err
}

func Exists(path string) bool {
	var err error

	_, err = os.Stat(path)

	if err == nil {
		return true
	}

	return false
}
