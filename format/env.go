package format

type FormatEnv struct {
	valFormatters map[byte]func()IValueFormatter
	exprFormatterConfig *ExprFormatterConfig
}

var env *FormatEnv = &FormatEnv{
	valFormatters: make(map[byte]func()IValueFormatter),
	exprFormatterConfig: NewExprFormatterConfig(),
}

//RegisterFormatter 注册一个格式化器
//@params key 格式化器的标签，单个字符; getFormatter 格式化器的工厂函数
func RegisterFormatter(key byte, getFormatter func()IValueFormatter) {
	env.valFormatters[key] = getFormatter
}

//RegisterInterpreter 注册一个表达式解释器
func RegisterInterpreter(key string, interpreter IExprInterpreter) {
	env.exprFormatterConfig.Interpreters[key] = interpreter
}

//SetDefaultInterpreter 设置默认的表达式解释器
func SetDefaultInterpreter(key string) {
	env.exprFormatterConfig.DefaultInter = key
}

//SetOperate 添加自定义操作符（默认自带+操作符，直接连接字符串）
func SetOperate(key byte, fn func(s1, s2 string) string) {
	env.exprFormatterConfig.BinOps[key] = fn
}

func init() {
	RegisterFormatter(STD_FORMATTER_LABEL, NewStdFormatter)
	RegisterFormatter(TIME_FORMATTER_LABEL, NewTimeFormatter)
	RegisterFormatter(PASSWORD_FORMAT_LABEL, NewPasswordFormatter)
}