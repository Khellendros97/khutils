package format

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	PASSWORD_FORMAT_LABEL = '*'
)

type PasswordFormatter struct {
	placeholder byte
	length      int
}

func NewPasswordFormatter() IValueFormatter {
	return &PasswordFormatter{}
}

func (f *PasswordFormatter) Parse(token string) (err error) {
	if len(token) <= 0 {
		f.placeholder = '*'
		return
	}
	f.placeholder = token[0]
	if len(token) > 1 {
		f.length, err = strconv.Atoi(token[1:])
	}
	return
}

func (f *PasswordFormatter) Format(value any) string {
	str := fmt.Sprintf("%v", value)
	if f.length <= 0 {
		f.length = len(str)
	}
	var sb strings.Builder
	for i := 0; i < f.length; i++ {
		sb.WriteByte(f.placeholder)
	}
	return sb.String()
}