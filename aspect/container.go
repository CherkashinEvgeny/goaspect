package aspect

import (
	"reflect"
)

type Container struct {
	aspects []Aspect
}

func (c *Container) Register(aspect Aspect) {
	c.aspects = append(c.aspects, aspect)
}

func (c *Container) Handler(ttype reflect.Type, method reflect.Method) Handler {
	aspect := &containerHandler{
		handlers: make([]Handler, 0, len(c.aspects)),
	}
	for _, handler := range c.aspects {
		aspect.handlers = append(aspect.handlers, handler.Handler(ttype, method))
	}
	return aspect
}

type containerHandler struct {
	handlers []Handler
}

func (c *containerHandler) Before(params ...any) {
	for _, handler := range c.handlers {
		handler.Before(params...)
	}
}

func (c *containerHandler) After(params ...any) {
	for _, handler := range c.handlers {
		handler.After(params...)
	}
}
