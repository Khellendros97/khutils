package format

import "fmt"

const STD_FORMATTER_LABEL = '%'

type StdFormatter struct {
	formatter string
}

func NewStdFormatter() IValueFormatter {
	return &StdFormatter{}
}

func (f *StdFormatter) Parse(token string) (err error) {
	f.formatter = token
	//fmt.Println("formatter", f.formatter)
	return
}

func (f *StdFormatter) Format(value any) string {
	return fmt.Sprintf("%" + f.formatter, value)
}

func (f *StdFormatter) ArgCount() int {
	return 1
}