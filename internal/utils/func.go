package utils

import (
	"reflect"
)

func NewFunction(fn interface{}) *Function {
	return &Function{
		fnType:  reflect.TypeOf(fn),
		fnValue: reflect.ValueOf(fn),
	}
}

type Function struct {
	fnType  reflect.Type
	fnValue reflect.Value
}

// An empty value is assigned when the missing parameter calls a function
func (f *Function) args(args ...interface{}) []reflect.Value {
	injmap := make(map[reflect.Type]reflect.Value)
	for i := range args {
		injmap[reflect.TypeOf(args[i])] = reflect.ValueOf(args[i])
	}
	count := f.fnType.NumIn()
	inValues := make([]reflect.Value, count)
	for i := 0; i < count; i++ {
		val, ok := injmap[f.fnType.In(i)]
		if ok {
			inValues[i] = val
			continue
		}
		inValues[i] = reflect.Zero(f.fnType.In(i))
	}
	return inValues
}

func (f *Function) IsFunc() bool {
	return f.fnType.Kind() == reflect.Func
}

func (f *Function) Invoke(fnArgs ...interface{}) []reflect.Value {
	if !f.IsFunc() {
		return []reflect.Value{}
	}
	return f.fnValue.Call(f.args(fnArgs...))
}
