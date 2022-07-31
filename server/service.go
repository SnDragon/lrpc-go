package server

import (
	"fmt"
	"go/ast"
	"reflect"
	"sync/atomic"
)

type methodType struct {
	method    reflect.Method
	argType   reflect.Type
	replyType reflect.Type
	numCall   uint64
}

func (m *methodType) newArgv() reflect.Value {
	var argv reflect.Value
	if m.argType.Kind() == reflect.Pointer {
		argv = reflect.New(m.argType.Elem())
	} else {
		argv = reflect.New(m.argType).Elem()
	}
	return argv
}

func (m *methodType) newReplyv() reflect.Value {
	replyv := reflect.New(m.replyType.Elem())
	switch m.replyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(m.replyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(m.replyType.Elem(), 0, 0))
	}
	return replyv
}

type service struct {
	name    string
	typ     reflect.Type
	rcvr    reflect.Value
	methods map[string]*methodType
}

func newService(rcvr interface{}) *service {
	s := &service{
		rcvr: reflect.ValueOf(rcvr),
		typ:  reflect.TypeOf(rcvr),
	}
	s.name = reflect.Indirect(s.rcvr).Type().Name()
	if !ast.IsExported(s.name) {
		panic(fmt.Sprintf("rpc server: %s is not a valid service name", s.name))
	}
	s.RegisterMethods()
	return s
}

func (s *service) RegisterMethods() {
	s.methods = make(map[string]*methodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		mType := method.Type
		/**
		RPC条件
		1. the method’s type is exported. – 方法所属类型是导出的。
		2. the method is exported. – 方法是导出的。
		3. the method has two arguments, both exported (or builtin) types. – 两个入参，均为导出或内置类型。
		4. the method’s second argument is a pointer. – 第二个入参必须是一个指针。
		5. the method has return type error. – 返回值为 error 类型。
		*/
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		argType, replyType := mType.In(1), mType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}
		s.methods[method.Name] = &methodType{
			method:    method,
			argType:   argType,
			replyType: replyType,
		}
		fmt.Printf("rpc server : %s registered\n", s.name)
	}
}

func (s *service) call(m *methodType, args, reply reflect.Value) error {
	atomic.AddUint64(&m.numCall, 1)
	f := m.method.Func
	returnValues := f.Call([]reflect.Value{s.rcvr, args, reply})
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}
