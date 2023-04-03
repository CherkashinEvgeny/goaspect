package aspect

import "reflect"

type Aspect interface {
	Handler(ttype reflect.Type, method reflect.Method) Handler
}

type Handler interface {
	Before(in ...any)
	After(out ...any)
}
