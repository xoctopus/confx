package mq

import (
	"container/list"
	"errors"
	"sync"
)

type Resource interface {
	// Elem returns element in list when removing
	Elem() *list.Element
	// SetElem set element in list when inserting
	SetElem(*list.Element)
}

type ReleaseOption struct {
	Unsub bool
}

type ReleaseOptionFunc func(*ReleaseOption)

type CanBeReleased interface {
	// Release close and cleanup resource. unsub denotes if unsubscribes before close
	Release(...ReleaseOptionFunc) error
}

func WithUnsubscribe() ReleaseOptionFunc {
	return func(o *ReleaseOption) { o.Unsub = true }
}

type manager struct {
	lst list.List
	mtx sync.RWMutex
}

func (r *manager) Add(v Resource) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	v.SetElem(r.lst.PushBack(v))
}

func (r *manager) Remove(v Resource, options ...ReleaseOptionFunc) error {
	defer func() {
		r.mtx.Lock()
		defer r.mtx.Unlock()
		r.lst.Remove(v.Elem())
	}()

	if x, ok := any(v).(CanBeReleased); ok {
		return x.Release(options...)
	}
	return nil
}

func (r *manager) Close() error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	errs := make([]error, 0, r.lst.Len())
	for r.lst.Len() > 0 {
		if e := r.lst.Front(); e != nil {
			r.lst.Remove(e)
			if x, ok := e.Value.(CanBeReleased); ok {
				errs = append(errs, x.Release())
			}
		}
	}
	return errors.Join(errs...)
}

func (r *manager) Len() int {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	return r.lst.Len()
}

// ResourceManager maintains all Consumer, Producer and Observer created by PubSub
type ResourceManager interface {
	ConsumerCount() int
	ProducerCount() int
	ObserverCount() int

	AddConsumer(Resource)
	AddProducer(Resource)
	AddObserver(Resource)

	Unsubscribe(Resource) error
	CloseConsumer(Resource) error
	CloseProducer(Resource) error
	CloseObserver(Resource) error
	Close() error
}

func NewResourceManager() ResourceManager {
	return &resources{
		consumers: &manager{},
		producers: &manager{},
		observers: &manager{},
	}
}

type resources struct {
	consumers *manager
	producers *manager
	observers *manager
}

func (m *resources) AddConsumer(r Resource) {
	m.consumers.Add(r)
}

func (m *resources) AddObserver(r Resource) {
	m.observers.Add(r)
}

func (m *resources) AddProducer(r Resource) {
	m.producers.Add(r)
}

func (m *resources) CloseConsumer(r Resource) error {
	return m.consumers.Remove(r)
}

func (m *resources) ConsumerCount() int {
	return m.consumers.Len()
}

func (m *resources) ProducerCount() int {
	return m.producers.Len()
}

func (m *resources) ObserverCount() int {
	return m.observers.Len()
}

func (m *resources) Unsubscribe(r Resource) error {
	return m.consumers.Remove(r, WithUnsubscribe())
}

func (m *resources) CloseObserver(r Resource) error {
	return m.observers.Remove(r)
}

func (m *resources) CloseProducer(r Resource) error {
	return m.producers.Remove(r)
}

func (m *resources) Close() error {
	return errors.Join(
		m.consumers.Close(),
		m.producers.Close(),
		m.observers.Close(),
	)
}
