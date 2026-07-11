package goruntime

import (
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
)

var corrMap sync.Map

func Goid() int64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, _ := strconv.ParseInt(idField, 10, 64)
	return id
}

func GetCorelationID() uuid.UUID {
	if v, ok := corrMap.Load(Goid()); ok {
		return v.(uuid.UUID)
	}
	id, _ := uuid.NewV7()
	corrMap.Store(Goid(), id)
	return id
}

func SetCorelationID(id uuid.UUID) {
	corrMap.Store(Goid(), id)
}

func ClearCorelationID() {
	corrMap.Delete(Goid())
}
