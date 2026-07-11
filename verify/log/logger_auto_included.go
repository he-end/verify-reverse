package log

import (
	"sync"

	"github.com/he-end/verify-reverse/verify/goruntime"
)

type RegisterRuntime struct {
	Key   string
	Value string
}

var runtimeStore sync.Map

func NewLoggerOnRuntime(reg RegisterRuntime) {
	runtimeStore.Store(goruntime.Goid(), &reg)
}

func DeferDeleteRuntimeValue() {
	runtimeStore.Delete(goruntime.Goid())
}

func GetLoggerRuntimeStore() *RegisterRuntime {
	if v, ok := runtimeStore.Load(goruntime.Goid()); ok {
		return v.(*RegisterRuntime)
	}
	return nil
}
