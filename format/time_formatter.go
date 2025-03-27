package format

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	TIME_FORMATTER_LABEL  = '@'
	TIME_FORMATTER_YMDHMS = "2006-01-02 15:04:05"
	TIME_FORMATTER_YMD    = "2006-01-02"
	TIME_FORMATTER_HMS    = "15:04:05"
)

type TimeFormatter struct {
	formatStr string
}

func NewTimeFormatter() IValueFormatter {
	return &TimeFormatter{}
}

func (f *TimeFormatter) Parse(token string) (err error) {
	switch token {
	case "datetime", "YMDhms":
		f.formatStr = TIME_FORMATTER_YMDHMS
	case "date", "YMD":
		f.formatStr = TIME_FORMATTER_YMD
	case "time", "hms":
		f.formatStr = TIME_FORMATTER_HMS
	default: {
		fs := strings.Replace(token, "Y", "2006", -1)
		fs = strings.Replace(fs, "M", "01", -1)
		fs = strings.Replace(fs, "D", "02", -1)
		fs = strings.Replace(fs, "h", "15", -1)
		fs = strings.Replace(fs, "m", "04", -1)
		fs = strings.Replace(fs, "s", "05", -1)
		f.formatStr = fs
	}
	}
	return
}

func (f *TimeFormatter) Format(value any) string {
	if t, ok := value.(time.Time); ok {
		return t.Format(f.formatStr)
	}
	var s string
	var ok bool
	if s, ok = value.(string); !ok {
		s = fmt.Sprintf("%v", value)
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err.Error()
	}
	return time.Unix(n, 0).Format(f.formatStr)
}
