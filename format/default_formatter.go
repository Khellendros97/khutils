package format

import "fmt"

type DefaultFormatter struct {
}

func NewDefaultFormatter() IValueFormatter {
	return &DefaultFormatter{}
}

func (f *DefaultFormatter) Parse(_ string) (err error) {
	return
}

func (f *DefaultFormatter) Format(value any) string {
	return fmt.Sprintf("%v", value)
}

func (f *DefaultFormatter) ArgCount() int {
	return 0
}
