package format

import (
	"fmt"
	"strings"
)

const LANG_FORMATTER_BEGIN = '$'

type LangFormatter struct {
	path []string
	Lang func(args ...string) string
}

func NewLangFormatter(lang func(args ...string) string) IValueFormatter {
	return &LangFormatter{
		Lang: lang,
	}
}

func (f *LangFormatter) Parse(token string) (err error) {
	f.path = strings.Split(token, ".")
	fmt.Println(token, f.path)
	if len(f.path) == 0 {
		return fmt.Errorf("Invalid lang formatter: %s", token)
	}
	return
}

func (f *LangFormatter) Format(value any) string {
	fmt.Println("Lang Format")
	return f.Lang(f.path...)
}

func (f *LangFormatter) ArgCount() int {
	return 0
}