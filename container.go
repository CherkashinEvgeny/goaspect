package aspect

import (
	"reflect"
	"sync"
)

type Container struct {
	underlying []Factory
}

func (c *Container) Register(factory Factory) {
	c.underlying = append(c.underlying, factory)
}

func (c *Container) Aspect(ttype reflect.Type, method reflect.Method) Aspect {
	aspect := aspectPool.Get().(*containerAspect)
	aspect.underlying = aspect.underlying[:0]
	for _, handler := range c.underlying {
		aspect.underlying = append(aspect.underlying, handler.Aspect(ttype, method))
	}
	return aspect
}

var aspectPool = &sync.Pool{
	New: func() any {
		return &containerAspect{
			underlying: make([]Aspect, 0, 10),
		}
	},
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
	aspectPool.Put(c)
}
