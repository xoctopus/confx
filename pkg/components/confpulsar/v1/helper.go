package confpulsar

import (
	"container/list"
	"sync"
)

type Elem interface {
	Elem() *list.Element
	SetElem(*list.Element)
	close()
}

type manager[T Elem] struct {
	lst list.List
	mtx sync.RWMutex
}

func (r *manager[T]) Add(v T) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	elem := r.lst.PushBack(v)
	v.SetElem(elem)
}

func (r *manager[T]) Remove(v T) {
	r.mtx.Lock()
	r.lst.Remove(v.Elem())
	r.mtx.Unlock()

	v.close()
}

func (r *manager[T]) Close() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	for e := r.lst.Front(); e != nil; e = e.Next() {
		v := e.Value.(T)
		v.close()
	}
}
