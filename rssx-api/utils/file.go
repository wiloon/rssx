package utils

import (
	"os"
	"rssx/utils/logger"
)

func IsFileOrDirExists(path string) bool {

	_, err := os.Stat(path)
	fileExist := err == nil || os.IsExist(err)

	logger.Infof("file: %s, exist:%v", path, fileExist)

	return fileExist
}
