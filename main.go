package main

import (
	"fmt"
	"time"

	//"strings"

	"github.com/Khellendros97/khutils/format"
)

type TestInterpreter struct {

}

func (i *TestInterpreter) Format(key string, args []any) (string, error) {
	return fmt.Sprintf("%s-%v", key, args), nil
}

func CN(key string) string {
	lang := map[string]string {
		"hello": "你好",
		"world": "世界",
		"pat": "值0: %s  值1: %s",
	}
	return lang[key]
}

func US(key string) string {
	lang := map[string]string {
		"hello": "hello",
		"world": "world",
		"pat": "val0: %s  val1: %s",
	}
	return lang[key]
}

var Lang string

type LangInterpreter struct {

}

func (i *LangInterpreter) Format(key string, args []any) (string ,error) {
	if Lang == "CN" {
		return fmt.Sprintf(CN(key), args...), nil
	} else if Lang == "US" {
		return fmt.Sprintf(US(key), args...), nil
	}
	return "", fmt.Errorf("invalid language")
}

func SetConcat() {
	if Lang == "CN" {
		format.SetOperate('+', func (s1, s2 string) string {
			return s1 + s2
		})
	} else if Lang == "US" {
		format.SetOperate('+', func(s1, s2 string) string {
			return s1 + " " + s2
		})
	}
}

func main() {
	Lang = "US"
	SetConcat()
	format.RegisterInterpreter("Lang", &LangInterpreter{})
	format.SetDefaultInterpreter("Lang")
	str := format.Fmt("{:*?6} {:@datetime}", "hyc1997115", time.Now())
	fmt.Println(str)
}