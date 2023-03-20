package aspect

import (
	"reflect"
)

type Container struct {
	underlying []Factory
}

func (c *Container) Register(factory Factory) {
	c.underlying = append(c.underlying, factory)
}

func (c *Container) Aspect(ttype reflect.Type, method reflect.Method) Aspect {
	aspect := &containerAspect{
		underlying: make([]Aspect, 0, len(c.underlying)),
	}
	for _, handler := range c.underlying {
		aspect.underlying = append(aspect.underlying, handler.Aspect(ttype, method))
	}
	return aspect
}

type containerAspect struct {
	underlying []Aspect
}

func (c *containerAspect) Before(params ...Param) {
	for _, aspect := range c.underlying {
		aspect.Before(params...)
	}
}

func (c *containerAspect) After(params ...Param) {
	for _, aspect := range c.underlying {
		aspect.After(params...)
	}
}
