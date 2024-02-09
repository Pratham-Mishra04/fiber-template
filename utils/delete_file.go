package utils

import (
	"os"
)

func DeleteFile(path string, fileName string) error {
	if fileName == "" || fileName == "default.jpg" {
		return nil
	}

	picPath := "public/" + path + "/" + fileName

	err := os.Remove(picPath)
	if err != nil {
		return err
	}

	return nil
}
