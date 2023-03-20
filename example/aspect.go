// Code generated by github.com/CherkashinEvgeny/goaspect. DO NOT EDIT.
package main

import (
	"context"
	"github.com/CherkashinEvgeny/goaspect"
	"reflect"
)

var mapType = reflect.TypeOf((*Map)(nil)).Elem()

type MapAspect struct {
	Impl    Map
	Factory aspect.Factory
}

var mapTypeMethodDelete, _ = mapType.MethodByName("Delete")

func (a MapAspect) Delete(arg0 context.Context, arg1 string) error {
	asp := a.Factory.Aspect(mapType, mapTypeMethodDelete)
	asp.Before(aspect.Param{
		Name:  "ctx",
		Value: arg0,
	}, aspect.Param{
		Name:  "key",
		Value: arg1,
	})
	res0 := a.Impl.Delete(arg0, arg1)
	asp.After(aspect.Param{
		Name:  "err",
		Value: res0,
	})
	return res0
}

var mapTypeMethodGet, _ = mapType.MethodByName("Get")

func (a MapAspect) Get(arg0 context.Context, arg1 string) ([]byte, error) {
	asp := a.Factory.Aspect(mapType, mapTypeMethodGet)
	asp.Before(aspect.Param{
		Name:  "ctx",
		Value: arg0,
	}, aspect.Param{
		Name:  "key",
		Value: arg1,
	})
	res0, res1 := a.Impl.Get(arg0, arg1)
	asp.After(aspect.Param{
		Name:  "value",
		Value: res0,
	}, aspect.Param{
		Name:  "err",
		Value: res1,
	})
	return res0, res1
}

var mapTypeMethodSet, _ = mapType.MethodByName("Set")

func (a MapAspect) Set(arg0 context.Context, arg1 string, arg2 []byte) error {
	asp := a.Factory.Aspect(mapType, mapTypeMethodSet)
	asp.Before(aspect.Param{
		Name:  "ctx",
		Value: arg0,
	}, aspect.Param{
		Name:  "key",
		Value: arg1,
	}, aspect.Param{
		Name:  "value",
		Value: arg2,
	})
	res0 := a.Impl.Set(arg0, arg1, arg2)
	asp.After(aspect.Param{
		Name:  "err",
		Value: res0,
	})
	return res0
}
