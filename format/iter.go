package format

import (
	"fmt"
	"strings"
)

type FormatState int

const (
	FORMAT_STATE_START             FormatState = iota
	FORMAT_STATE_LITERAL                       // 原样输出的字面量
	FORMAT_STATE_PLACEHOLDER_START             // 解析到'{'字符
	FORMAT_STATE_PARSE_INDEX                   // '{'后紧跟一个数字，表示参数索引
	FORMAT_STATE_PARSE_FORMATTER               // ':'后跟着一个格式化器
	FORMAT_STATE_PLACEHOLDER_END               // 解析到'}'字符
	FORMAT_STATE_EXPR                          // 解析到{{表达式
	FORMAT_STATE_EXPR_END                      // 解析到}}表达式终止
	FORMAT_STATE_ERROR                         // 解析错误
	FORMAT_STATE_END
)

func (s FormatState) String() string {
	switch s {
	case FORMAT_STATE_START:
		return "FORMAT_STATE_START"
	case FORMAT_STATE_LITERAL:
		return "FORMAT_STATE_LITERAL"
	case FORMAT_STATE_PLACEHOLDER_START:
		return "FORMAT_STATE_PLACEHOLDER_START"
	case FORMAT_STATE_PARSE_INDEX:
		return "FORMAT_STATE_PARSE_INDEX"
	case FORMAT_STATE_PARSE_FORMATTER:
		return "FORMAT_STATE_PARSE_FORMATTER"
	case FORMAT_STATE_PLACEHOLDER_END:
		return "FORMAT_STATE_PLACEHOLDER_END"
	case FORMAT_STATE_END:
		return "FORMAT_STATE_END"
	case FORMAT_STATE_ERROR:
		return "FORMAT_STATE_ERROR"
	case FORMAT_STATE_EXPR:
		return "FORMAT_STATE_EXPR"
	case FORMAT_STATE_EXPR_END:
		return "FORMAT_STATE_EXPR_END"
	default:
		return "UNKNOWN_FORMAT_STATE"
	}
}

func (s FormatState) startStateNext(ch byte, pos *int) FormatState {
	if ch == '{' { // 如果遇到'{'字符，进入占位符解析状态
		(*pos)++
		return FORMAT_STATE_PLACEHOLDER_START
	}
	// 否则进入字面量状态
	return FORMAT_STATE_LITERAL
}

func (s FormatState) literalStateNext(ch byte, pos *int) FormatState {
	(*pos)++
	if ch == '{' { // 如果遇到'{'字符，进入占位符解析状态
		return FORMAT_STATE_PLACEHOLDER_START
	}
	// 否则继续保持字面量状态
	return FORMAT_STATE_LITERAL
}

func (s FormatState) placeholderStartStateNext(ch byte, pos *int) FormatState {
	//fmt.Println("placeholderStartStateNext ", string([]byte{ch}))
	if ch >= '0' && ch <= '9' { // 如果遇到数字字符，进入参数索引解析状态
		return FORMAT_STATE_PARSE_INDEX
	} else if ch == ':' { // 如果遇到':'字符，进入格式化器解析状态
		(*pos)++
		return FORMAT_STATE_PARSE_FORMATTER
	} else if ch == '}' { // 如果遇到'}'字符，进入占位符结束状态
		(*pos)++
		return FORMAT_STATE_PLACEHOLDER_END
	} else if ch == '{' { // 双层{表示表达式
		(*pos)++
		return FORMAT_STATE_EXPR
	}
	// 否则返回错误状态
	return FORMAT_STATE_ERROR
}

func (s FormatState) exprStateNext(ch byte, pos *int) FormatState {
	(*pos)++
	if ch == '}' {
		return FORMAT_STATE_EXPR_END
	}
	return FORMAT_STATE_EXPR
}

func (s FormatState) exprStateEndNext(ch byte, pos *int) FormatState {
	if ch == '}' {
		(*pos)++
		return FORMAT_STATE_PLACEHOLDER_END
	}
	return FORMAT_STATE_ERROR
}

func (s FormatState) parseIndexStateNext(ch byte, pos *int) FormatState {
	(*pos)++
	if ch >= '0' && ch <= '9' { // 如果遇到数字字符，继续保持参数索引解析状态
		return FORMAT_STATE_PARSE_INDEX
	} else if ch == ':' { // 如果遇到':'字符，进入格式化器解析状态
		return FORMAT_STATE_PARSE_FORMATTER
	} else if ch == '}' { // 如果遇到'}'字符，进入占位符结束状态
		return FORMAT_STATE_PLACEHOLDER_END
	}
	// 否则返回错误状态
	return FORMAT_STATE_END
}

func (s FormatState) parseFormatterStateNext(ch byte, pos *int) FormatState {
	(*pos)++
	if ch == '}' { // 如果遇到'}'字符，进入占位符结束状态
		return FORMAT_STATE_PLACEHOLDER_END
	}
	// 否则继续保持格式化器解析状态
	return FORMAT_STATE_PARSE_FORMATTER
}

func (s FormatState) placeholderEndStateNext(ch byte, pos *int) FormatState {
	return FORMAT_STATE_START
}

func (s FormatState) Next(ch byte, pos *int) FormatState {
	if ch == 0 {
		return FORMAT_STATE_END
	}
	switch s {
	case FORMAT_STATE_START:
		return s.startStateNext(ch, pos)
	case FORMAT_STATE_LITERAL:
		return s.literalStateNext(ch, pos)
	case FORMAT_STATE_PLACEHOLDER_START:
		return s.placeholderStartStateNext(ch, pos)
	case FORMAT_STATE_PARSE_INDEX:
		return s.parseIndexStateNext(ch, pos)
	case FORMAT_STATE_PARSE_FORMATTER:
		return s.parseFormatterStateNext(ch, pos)
	case FORMAT_STATE_PLACEHOLDER_END:
		return s.placeholderEndStateNext(ch, pos)
	case FORMAT_STATE_EXPR:
		return s.exprStateNext(ch, pos)
	case FORMAT_STATE_EXPR_END:
		return s.exprStateEndNext(ch, pos)
	default:
		return FORMAT_STATE_ERROR
	}
}

type FormatIter struct {
	input []byte
	pos   int
	state FormatState
}

func NewFormatIter(input string) *FormatIter {
	return &FormatIter{input: []byte(input), pos: 0, state: FORMAT_STATE_START}
}

func (i *FormatIter) Next() (FormatState, byte, error) {
	//fmt.Println("pos: ", i.pos)
	//fmt.Printf("state: %s, pos: %d\n", i.state.String(), i.pos)
	if i.pos >= len(i.input) {
		return FORMAT_STATE_END, 0, IterEndError{}
	}
	ch := i.input[i.pos]
	i.state = i.state.Next(ch, &i.pos)
	return i.state, ch, nil
}

func (i *FormatIter) NextToken() (FormatState, string, error) {
	//fmt.Println("NextToken")
	var sb strings.Builder
	originState := i.state
	for {
		state, ch, err := i.Next()
		//fmt.Printf("pos: %d, ch: %c\n", i.pos, ch)
		if err != nil { // 如果读取到字符串结尾，返回迭代器当前状态和读取到的字符串
			return state, sb.String(), err
		}
		if state == FORMAT_STATE_ERROR { // 如果读取到错误状态，返回错误
			return state, sb.String(), fmt.Errorf("Invalid format string: %s", sb.String())
		}
		if state != originState { // 如果迭代器状态产生变化，返回迭代器当前状态和读取到的字符串
			//fmt.Printf("token: [%s], next state: [%s]\n", sb.String(), state.String())
			return state, sb.String(), nil
		}
		sb.WriteByte(ch)
	}
}

func (i *FormatIter) GetState() FormatState {
	return i.state
}
