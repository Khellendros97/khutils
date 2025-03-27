package format

import (
	"fmt"
	"strconv"
	"unicode"
)

type Expr interface {
	Eval(*ExprFormatter) (string, error)
}

type tokenOp byte

type tokenLiteral string

func (l *tokenLiteral) Eval(env *ExprFormatter) (string, error) {
	return string(*l), nil
}

type tokenLabel string
type tokenVar struct {
	namespace tokenLabel
	key       tokenLabel
}

func (l *tokenVar) Eval(env *ExprFormatter) (string, error) {
	return env.EvalVar(string(l.namespace), string(l.key), []any{})
}

type tokenParam int

type tokenFunc struct {
	namespace tokenLabel
	key       tokenLabel
	params    []tokenParam
}

func (l *tokenFunc) Eval(env *ExprFormatter) (str string, err error) {
	args := make([]any, len(l.params))
	for i, param := range l.params {
		args[i], err = env.GetArg(int(param))
		if err != nil {
			return
		}
	}
	return env.EvalVar(string(l.namespace), string(l.key), args)
}

type binaryExpr struct {
	left  Expr
	right Expr
	op    tokenOp
}

func (e *binaryExpr) Eval(env *ExprFormatter) (str string, err error) {
	fn, ok := env.BinOps[byte(e.op)]
	if !ok {
		return "", fmt.Errorf("unknown operator: %c", e.op)
	}
	vl, err := e.left.Eval(env)
	if err != nil {
		return "", err
	}
	vr, err := e.right.Eval(env)
	if err != nil {
		return "", err
	}
	str = fn(vl, vr)
	return
}

type ExprParser struct {
	expr string
	pos  int
}

func NewExprParser(expr string) *ExprParser {
	return &ExprParser{expr: expr}
}

func (p *ExprParser) next() (ch byte, err error) {
	//fmt.Printf("next: ")
	if p.pos >= len(p.expr) {
		return 0, IterEndError{}
	}
	ch = p.expr[p.pos]
	//fmt.Printf("%c\n", ch)
	p.pos++
	if ch == ' ' {
		return p.next()
	}
	return
}

func (p *ExprParser) skipSpace() {
	for p.pos < len(p.expr) {
		if p.expr[p.pos] == ' ' {
			p.pos++
		} else {
			return
		}
	}
}

func (p *ExprParser) current() (ch byte, err error) {
	if p.pos >= len(p.expr) {
		return 0, IterEndError{}
	}
	ch = p.expr[p.pos]
	return
}

func (p *ExprParser) getPos() int {
	p.current()
	return p.pos
}

func (p *ExprParser) Expect(pred func(byte)bool) (bool, error) {
	if p.pos >= len(p.expr) {
		return false, IterEndError{}
	}
	ch, err := p.next()
	if err != nil {
		return false, err
	}

	if pred(ch)  {
		return true, nil
	}
	return false, nil
}

func (p *ExprParser) ExpectString(str string) (bool, error) {
	for _, ch := range []byte(str) {
		ok, err := p.Expect(isChar(ch))
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

func (p *ExprParser) Advance(step int) error {
	for i := 0; i < step; i++ {
		_, err := p.next()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *ExprParser) Require(pred func(byte)bool) (bool, error) {
	if p.pos >= len(p.expr) {
		return false, IterEndError{}
	}
	ch, err := p.current()
	if err != nil {
		return false, err
	}
	if pred(ch) {
		return true, nil
	}
	return false, nil
}

func (p *ExprParser) RequireString(str string) (bool, error) {
	pos := p.pos
	ok, err := p.ExpectString(str)
	p.pos = pos
	return ok, err
}

func isOp(ch byte) bool {
	switch ch {
	case '+': return true
	default: return false
	}
}

func isChar(ch byte) func (byte) bool {
	return func (ch2 byte) bool {
		return ch == ch2
	}
}

func (p *ExprParser) residue() string {
	if p.pos < len(p.expr) {
		return p.expr[p.pos:]
	}
	return ""
}

func (p *ExprParser) ParseExpr() (Expr, error) {
	//fmt.Println(">>> expr", p.residue())
	if p.pos >= len(p.expr) {
		return nil, IterEndError{} // End of iteration
	}
	p.skipSpace()
	pos := p.pos
	be, err := p.parseBinaryExpr()
	if err == nil || IsIterEnd(err) {
		return be, nil
	} else {
		//fmt.Println(err)
		p.pos = pos
	}
	fn, err := p.parseFunc()
	if err == nil || IsIterEnd(err) {
		return fn, nil
	} else {
		//fmt.Println(err)
		p.pos = pos
	}
	variable, err := p.parseVariable()
	if err == nil || IsIterEnd(err) {
		return variable, nil
	} else {
		//fmt.Println(err)
		p.pos = pos
	}
	literal, err := p.parseLiteral()
	if err == nil || IsIterEnd(err) {
		return literal, nil
	}
	//fmt.Println(err)
	return nil, err
}

func (p *ExprParser) ParseUnitExpr() (Expr, error) {
	//fmt.Println(">>> unit expr", p.residue())
	p.skipSpace()
	pos := p.pos
	if p.pos >= len(p.expr) {
		return nil, IterEndError{} // End of iteration
	}
	fn, err := p.parseFunc()
	if err == nil || IsIterEnd(err) {
		return fn, nil
	} else {
		p.pos = pos
	}
	variable, err := p.parseVariable()
	if err == nil || IsIterEnd(err) {
		return variable, nil
	} else {
		p.pos = pos
	}
	literal, err := p.parseLiteral()
	if err == nil || IsIterEnd(err) {
		return literal, nil
	}
	return nil, err
}

func (p *ExprParser) parseBinaryExpr() (*binaryExpr, error) {
	//fmt.Println(">>> binary", p.residue())
	left, err := p.ParseUnitExpr()
	if err != nil {
		return nil, err
	}
	p.skipSpace()
	isop, err := p.Require(isOp)
	if err != nil {
		return nil, fmt.Errorf("missing operator")
	}

	if !isop {
		return nil, fmt.Errorf("invalid operator: %c")
	}
	op, _ := p.current()
	p.Advance(1)

	right, err := p.ParseExpr()
	if err != nil {
		return nil, err
	}

	return &binaryExpr{left: left, right: right, op: tokenOp(op)}, nil
}

func (p *ExprParser) parseFunc() (*tokenFunc, error) {
	//fmt.Println(">>> func", p.residue())
	v, err := p.parseName()
	if err != nil {
		//fmt.Println(err)
		return nil, err
	}
	ok, err := p.Expect(isChar('('))
	if err != nil {
		//fmt.Println(err)
		return nil, fmt.Errorf("missing (")
	}
	if !ok {
		return nil, fmt.Errorf("missing '('")
	}
	params := make([]tokenParam, 0)
	for {
		param, err := p.parseParam()
		if err != nil {
			break
		}
		params = append(params, *param)
		ok, err := p.Require(isChar(','))
		if err != nil {
			return nil, fmt.Errorf("missing ','")
		}
		if !ok {
			break
		}
		p.Advance(1)
	}
	////fmt.Println("expect )")
	ok, err = p.Expect(isChar(')'))
	if err != nil {
		//fmt.Println(err)
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("missing ')'")
	}

	fn := &tokenFunc{namespace: v.namespace, key: v.key, params: params}
	//fmt.Println(fn)
	return fn, nil
}

func (p *ExprParser) parseVariable() (*tokenVar, error) {
	//fmt.Println(">>> var", p.residue())
	v, err := p.parseName()
	if err != nil {
		return nil, err
	}
	ok ,_ := p.Require(isChar('('))
	if ok {
		return nil, fmt.Errorf("unexpect '('")
	}
	return v, nil
}

func (p *ExprParser) parseName() (*tokenVar, error) {
	//fmt.Println(">>> name", p.residue())
	var namespace *tokenLabel
	var key *tokenLabel
	label, err := p.parseLabel()
	if err != nil {
		return nil, err
	}
	ok, err := p.RequireString("::")
	if err != nil && !IsIterEnd(err) {
		return nil, err
	}
	if ok {
		namespace = label
		p.Advance(2)
		label, err = p.parseLabel()
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("invalid expr: %s", p.expr)
		}
		key = label
	} else {
		namespace = new(tokenLabel)
		key = label
	}
	////fmt.Printf("namespace: %s, key: %s", *namespace, *key)
	return &tokenVar{namespace: *namespace, key: *key}, nil
}

func (p *ExprParser) parseLiteral() (*tokenLiteral, error) {
	//fmt.Println(">>> literal", p.residue())
	var value tokenLiteral
	ok, err := p.Expect(isChar('\''))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("missing '\\''")
	}
	pstart := p.getPos()
	for {
		ok, err := p.Require(isChar('\''))
		if err != nil {
			return nil, err
		}
		if ok {
			break
		}
		p.Advance(1)
	}
	value = tokenLiteral(p.expr[pstart:p.pos])
	ok, err = p.Expect(isChar('\''))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("missing '\\''")
	}
	return &value, nil
}

func (p *ExprParser) parseLabel() (*tokenLabel, error) {
	//fmt.Println(">>> label", p.residue())
	var value tokenLabel
	pstart := p.getPos()
	for {
		ok, err := p.Require(func (ch byte) bool {
			return unicode.IsLetter(rune(ch)) || unicode.IsNumber(rune(ch)) || ch == '_' || ch == '.'
		})
		if err != nil && !IsIterEnd(err){
			return nil, err
		}
		if !ok || IsIterEnd(err) {
			break
		}
		p.Advance(1)
	}
	value = tokenLabel(p.expr[pstart:p.pos])
	if len(value) == 0 {
		return nil, fmt.Errorf("empty label")
	}
	//fmt.Printf("label: [%s]\n", value)
	return &value, nil
}

func (p *ExprParser) parseParam() (*tokenParam, error) {
	//fmt.Println(">>> param", p.residue())
	var param tokenParam
	ok, err := p.Expect(isChar('$'))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("missing '$'")
	}
	pstart := p.getPos()
	for {
		ok, err := p.Require(func (ch byte) bool {
			return unicode.IsLetter(rune(ch)) || unicode.IsNumber(rune(ch))
		})
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		p.Advance(1)
	}

	str := p.expr[pstart:p.pos]
	index, err := strconv.Atoi(str)
	if err != nil {
		return nil, err
	}
	param = tokenParam(index)
	return &param, nil

}