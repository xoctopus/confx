package mq_test

import (
	"container/list"
	"fmt"

	"github.com/xoctopus/confx/pkg/types/mq"
)

type MockResource struct {
	name string
	idx  int
	elem *list.Element
}

func (r *MockResource) Elem() *list.Element {
	return r.elem
}

func (r *MockResource) SetElem(e *list.Element) {
	r.elem = e
}

func (r *MockResource) Release(appliers ...mq.ReleaseOptionFunc) error {
	opt := &mq.ReleaseOption{}
	for _, applier := range appliers {
		applier(opt)
	}

	if !opt.Unsub {
		fmt.Printf("\tresource %s:%d released\n", r.name, r.idx)
	} else {
		fmt.Printf("\tresource %s:%d unsubscribed and released\n", r.name, r.idx)
	}
	return nil
}

func ExampleResourceManager() {
	rm := mq.NewResourceManager()

	defer func() {
		fmt.Println("Close resource manager:")
		_ = rm.Close()
		fmt.Printf("\tproducers: %d\n", rm.ProducerCount())
		fmt.Printf("\tconsumers: %d\n", rm.ConsumerCount())
		fmt.Printf("\tobservers: %d\n", rm.ObserverCount())
	}()

	fmt.Println("Empty resource:")
	fmt.Printf("\tproducers: %d\n", rm.ProducerCount())
	fmt.Printf("\tconsumers: %d\n", rm.ConsumerCount())
	fmt.Printf("\tobservers: %d\n", rm.ObserverCount())

	rm.AddProducer(&MockResource{name: "producer", idx: 1})
	p := &MockResource{name: "producer", idx: 2}
	rm.AddProducer(p)
	rm.AddConsumer(&MockResource{name: "consumer", idx: 1})
	c1 := &MockResource{name: "consumer", idx: 2}
	rm.AddConsumer(c1)
	c2 := &MockResource{name: "consumer", idx: 3}
	rm.AddConsumer(c2)
	rm.AddObserver(&MockResource{name: "observer", idx: 1})
	o := &MockResource{name: "observer", idx: 2}
	rm.AddObserver(o)

	fmt.Println("After add resources:")
	fmt.Printf("\tproducers: %d\n", rm.ProducerCount())
	fmt.Printf("\tconsumers: %d\n", rm.ConsumerCount())
	fmt.Printf("\tobservers: %d\n", rm.ObserverCount())

	fmt.Println("Close or Unsubscribe resources:")
	_ = rm.CloseProducer(p)
	_ = rm.CloseConsumer(c1)
	_ = rm.Unsubscribe(c2)
	_ = rm.CloseObserver(o)

	fmt.Println("After releases:")
	fmt.Printf("\tproducers: %d\n", rm.ProducerCount())
	fmt.Printf("\tconsumers: %d\n", rm.ConsumerCount())
	fmt.Printf("\tobservers: %d\n", rm.ObserverCount())

	//Output:
	//Empty resource:
	//	producers: 0
	//	consumers: 0
	//	observers: 0
	//After add resources:
	//	producers: 2
	//	consumers: 3
	//	observers: 2
	//Close or Unsubscribe resources:
	//	resource producer:2 released
	//	resource consumer:2 released
	//	resource consumer:3 unsubscribed and released
	//	resource observer:2 released
	//After releases:
	//	producers: 1
	//	consumers: 1
	//	observers: 1
	//Close resource manager:
	//	resource consumer:1 released
	//	resource producer:1 released
	//	resource observer:1 released
	//	producers: 0
	//	consumers: 0
	//	observers: 0
}
