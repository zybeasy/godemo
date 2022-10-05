/* 使用反射输入任意值的类型并枚举它的方法 */
package main

import (
	"fmt"
	"reflect"
	"strings"
)

func PrintInfo(x interface{}) {
	fmt.Printf("\033[0;31m====> PrintInfo %s :\033[0m\n", reflect.TypeOf(x))
	v := reflect.ValueOf(x)
	t := v.Type()
	fmt.Printf("type %s\n", t)

	for i := 0; i < v.NumMethod(); i++ {
		methType := v.Method(i).Type()
		fmt.Printf("func (%s) %s%s\n", t, t.Method(i).Name,
			strings.TrimPrefix(methType.String(), "func"))
	}
}
