package ioc

import (
	"net/http"
	"reflect"
	"sync"
)

const (
	transient scope = iota
	singleton
)

var c *container = newContainer()

type (
	scope     int
	container struct {
		bindings map[string]*binding
		mu       sync.RWMutex
	}
	binding struct {
		constructor any
		instance    any
		scope       scope
	}
	Handler interface {
		Handle(w http.ResponseWriter, r *http.Request)
	}
)

func newContainer() *container {
	return &container{
		bindings: make(map[string]*binding),
	}
}

func bind(constructor any, scope scope) {

	constructorType := reflect.TypeOf(constructor)

	if constructorType.Kind() != reflect.Func {
		panic("Constructor must be a function")
	}

	if constructorType.NumOut() != 1 {
		panic("Constructor must return a single value")
	}

	typeName := constructorType.Out(0).String()

	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.bindings[typeName]; exists {
		panic("Type already bound: " + typeName)
	}

	c.bindings[typeName] = &binding{
		constructor: constructor,
		scope:       scope,
	}
}

func BindSingleton(constructor any) {
	bind(constructor, singleton)
}

func BindTransient(constructor any) {
	bind(constructor, transient)
}

func resolveByType(bindType reflect.Type) any {
	typeName := bindType.String()

	c.mu.RLock()
	defer c.mu.RUnlock()
	constructor, exists := c.bindings[typeName]
	if !exists {
		panic("Unbound type: " + typeName)
	}

	if constructor.scope == transient {
		return invoke(constructor.constructor)
	}

	if constructor.instance == nil {
		constructor.instance = invoke(constructor.constructor)
	}

	return constructor.instance
}

func invoke(constructor any) any {
	constructorValue := reflect.ValueOf(constructor)
	constuctorType := constructorValue.Type()

	var args []reflect.Value
	for i := 0; i < constuctorType.NumIn(); i++ {
		argType := constuctorType.In(i)
		argValue := reflect.ValueOf(resolveByType(argType))
		args = append(args, argValue)
	}
	results := constructorValue.Call(args)
	return results[0].Interface()
}

func Resolve[T any]() T {
	return resolveByType(reflect.TypeOf((*T)(nil)).Elem()).(T)
}

func ResolveHanlder[T Handler](w http.ResponseWriter, r *http.Request) {
	handler := resolveByType(reflect.TypeOf((*T)(nil)).Elem()).(T)
	handler.Handle(w, r)
}
