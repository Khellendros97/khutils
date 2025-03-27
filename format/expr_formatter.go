package format

import "fmt"

type ExprFormatterConfig struct {
	Interpreters map[string]IExprInterpreter
	DefaultInter string
	BinOps       map[byte]func(s1, s2 string) string
}

func NewExprFormatterConfig() *ExprFormatterConfig {
	return &ExprFormatterConfig{
		Interpreters: make(map[string]IExprInterpreter),
		BinOps: map[byte]func(s1, s2 string) string{
			'+': func(s1, s2 string) string {
				return s1 + s2
			},
		},
	}
}

type ExprFormatter struct {
	*ExprFormatterConfig
	Args         []any
}

func NewExprFormatter(config *ExprFormatterConfig) *ExprFormatter {
	return &ExprFormatter{
		ExprFormatterConfig: config,
	}
}

func (f *ExprFormatterConfig) Register(name string, interpreter IExprInterpreter) {
	f.Interpreters[name] = interpreter
}

func (f *ExprFormatterConfig) SetDefault(name string) {
	f.DefaultInter = name
}

func (f *ExprFormatterConfig) SetFnConcat(key byte, fn func(s1, s2 string) string) {
	f.BinOps[key] = fn
}

func (f *ExprFormatter) GetArg(index int) (any, error) {
	if index < len(f.Args) {
		return f.Args[index], nil
	}
	return nil, fmt.Errorf("Argument index out of range: %d", index)
}

func (f *ExprFormatter) EvalVar(namespace string, key string, args []any) (string, error) {
	if namespace == "" {
		namespace = f.DefaultInter
	}
	module := f.Interpreters[namespace]
	if module == nil {
		return "", fmt.Errorf("Unknown namespace: %s", namespace)
	}
	return module.Format(key, args)
}

func (f *ExprFormatter) Eval(expr string) (str string, err error) {
	parser := NewExprParser(expr)
	ex, err := parser.ParseExpr()
	if err != nil {
		return
	}
	//fmt.Println("eval", ex)
	return ex.Eval(f)
}
