package format

import (
	"fmt"
	"strconv"
	"strings"
)

type format struct {
	sb strings.Builder
	count int
	formatter IValueFormatter
	args []any
	lastState FormatState
	iter *FormatIter
	exprFormatter *ExprFormatter
}

func (f *format) formatValue(index int) (err error) {
	id := index
	if index < 0 {
		if f.count < 0 {
			err = fmt.Errorf("can not mix indexed param with non-indexed param")
			return
		}
		id = f.count
	} else {
		f.count = -1
	}
	if id >= len(f.args) {
		err = fmt.Errorf("arg index out of range")
		return
	}
	var str string
	if f.formatter == nil {
		str = fmt.Sprintf("%v", f.args[id])
	} else {
		str = f.formatter.Format(f.args[id])
	}
	
	f.sb.WriteString(str)
	if index < 0 {
		f.count++
	}
	return
}

func (f *format) format() string {
	index := -1
	var err error
	for {
		state, token, err := f.iter.NextToken()
		//fmt.Println("state: ", state, ", token: ", token)
		if err != nil && !IsIterEnd(err) {
			break
		}
		switch f.lastState {
		case FORMAT_STATE_LITERAL:
			f.sb.WriteString(token)
		case FORMAT_STATE_PARSE_INDEX:
			index, err = strconv.Atoi(token)
			if err != nil {
				break
			}
		case FORMAT_STATE_PARSE_FORMATTER:
			getFmt, ok := env.valFormatters[token[0]]
			if !ok {
				f.formatter = &DefaultFormatter{}
			}
			f.formatter = getFmt()
			f.formatter.Parse(token[1:])
			f.formatValue(index)
		case FORMAT_STATE_EXPR:
			var str string
			str, err = f.exprFormatter.Eval(token)
			if err != nil {
				break
			}
			f.sb.WriteString(str)
		case FORMAT_STATE_PLACEHOLDER_END:
			index = -1
			f.formatter = nil
		}

		if state == FORMAT_STATE_PLACEHOLDER_END && (f.lastState == FORMAT_STATE_PLACEHOLDER_START || f.lastState == FORMAT_STATE_PARSE_INDEX) {
			f.formatValue(index)
		}

		f.lastState = state
		if IsIterEnd(err) {
			break
		}
	}

	if IsIterEnd(err) {
		return f.sb.String()
	} else if err != nil {
		return err.Error()
	}
	return f.sb.String()
}

//Fmt 可自定义格式化器，以及插入表达式的格式化
//使用含参数索引的格式化器：
//Fmt("{1} {1}, {0}", "world", "hello") => "hello hello world"
//省略参数索引时，会按顺序使用参数
//Fmt("{}, {}", "hello", "world") => "hello world"
//在索引后面跟:<格式化器标签>来使用格式化器，默认自带的std格式化器的标签是%，用法与fmt.Printf类似：
//Fmt("{0:%.2f}", 3.1415926) => "3.14"
//也可以省略参数索引：
//Fmt("{:%.2f}", 3.1415926) => "3.14"
//假设你注册了一个自定义格式化器PasswordFormatter，标签为*，用于隐藏密码，那么你应该这么使用它：
//Fmt("{:*}", "password") => "********
//在你注册了表达式解析器之后，你可以在格式化字符串中插入表达式，这可以实现字符串翻译等功能
//Fmt("{{Lang::hello + world + ', ' + myNameIs($0)}}", "John") => "你好世界我是John"
//表达式需要使用双层大括号包裹。
//Lang是表达式解析器所在的命名空间，后面紧跟两个冒号，然后是系统变量或函数的名字，可以包含字母、数字、下划线和.
//如果你将Lang设置为默认解析器，那么你可以省略Lang::，直接写变量/函数的名字，例如将Lang::hello改为hello
//可以使用$0、$1、$2等来引用格式化参数，$0表示第一个参数，$1表示第二个参数，以此类推
//字符串常量需要使用单引号包裹
func Fmt(pattern string, args ...any) string {
	format := &format{
		args: args,
		iter: NewFormatIter(pattern),
		exprFormatter: NewExprFormatter(env.exprFormatterConfig),
	}

	format.exprFormatter.Args = args

	return format.format()
}