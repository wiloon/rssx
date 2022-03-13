package utils

import (
	"fmt"
	"rssx/utils/logger"
	"runtime"
)

func RecoverAndPrintStackTrace() {
	if err := recover(); err != nil {
		logger.Error("recover statistics, e: %v", err)
		buf := make([]byte, 2048)
		n := runtime.Stack(buf, false)
		stackInfo := fmt.Sprintf("%s", buf[:n])
		logger.Error("panic stack info %s", stackInfo)

	}
}
