package goaspect

import "reflect"

type Factory interface {
	Aspect(ttype reflect.Type, method *reflect.Method) Aspect
}

type Aspect interface {
	Before(inParams ...Param)
	After(outParams ...Param)
}

type Param struct {
	Name  string
	Value any
}
