package format

import "errors"

type IterEndError struct {
}

func (e IterEndError) Error() string {
	return "End of iteration"
}

type InvalidFormatterError struct {
	Formatter string
}

func (e InvalidFormatterError) Error() string {
	return "Invalid formatter: " + e.Formatter
}

func IsIterEnd(err error) bool {
	var e IterEndError
	if errors.As(err, &e) {
		return true
	}
	return false
}

type IFormatIter interface {
	Next() (FormatState, byte, error)        // 读取下一个字符并返回迭代器当前状态，读取到字符串结尾时返回IterEndError
	NextToken() (FormatState, string, error) // 读取字符直到迭代器状态产生变化，返回迭代器当前状态，读取到字符串结尾时返回IterEndError
	GetState() FormatState                   // 获取迭代器状态
}

//IValueFormatter 格式化器接口
type IValueFormatter interface {
	//Parse 解析格式化字符串（不包含格式化器标签，例如：{:%.2f}传入的token参数为.2f
	Parse(token string) error
	//Format 格式化
	Format(value any) string
}

//IExprInterpreter 表达式解释器接口
type IExprInterpreter interface {
	//Format 将传入的表达式（变量和函数）求值，返回字符串
	//@params key 变量名或函数名，例如{{name}}（传入的key为name）或{{fn($0)}}（传入的key为fn）
	//@params args 传入的实参
	Format(key string, args []any) (string, error)
}
