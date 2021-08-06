package utils

import (
	"runtime"
	"strconv"
	"strings"
)

func FileWithLineNum() string {
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)

		if ok && strings.Contains(file, "/yatai/api-server/") {
			return file + ":" + strconv.FormatInt(int64(line), 10)
		}
	}
	return ""
}
