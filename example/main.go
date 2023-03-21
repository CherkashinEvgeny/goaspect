package main

import (
	"context"
	"fmt"
	"reflect"

	aspect "github.com/CherkashinEvgeny/goaspect"
)

func main() {
	container := aspect.Container{}
	container.Register(Logger{})
	var m Map
	m = MapAspect{Impl: &LocalMap{}, Factory: &container}
	err := m.Set(context.Background(), "hehe", nil)
	if err != nil {
		return
	}
	_, err = m.Get(context.Background(), "hehe")
	if err != nil {
		return
	}
	err = m.Delete(context.Background(), "hehe")
	if err != nil {
		return
	}
}

type Logger struct {
}

func (l Logger) Aspect(ttype reflect.Type, method reflect.Method) aspect.Aspect {
	return operationLogger{ttype: ttype, method: method}
}

type operationLogger struct {
	ttype  reflect.Type
	method reflect.Method
}

func (o operationLogger) Before(inParams ...aspect.Param) {
	fmt.Println("Before", o.ttype.Name(), o.method.Name, inParams)
}

func (o operationLogger) After(outParams ...aspect.Param) {
	fmt.Println("After", o.ttype.Name(), o.method.Name, outParams)
}
