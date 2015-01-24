package exec

import (
	"sync"
	"sync/atomic"
)

func NewDestroyable(destroyCallback func() error) Destroyable {
	return &destroyable{
		destroyCallback,
		sync.NewCond(&sync.Mutex{}),
		newVolatileBool(false),
		0,
		make([]Destroyable, 0),
	}
}

type destroyable struct {
	destroyCallback func() error
	cv              *sync.Cond
	destroyed       *volatileBool
	numOperations   int32
	children        []Destroyable
}

func (this *destroyable) Destroy() error {
	this.cv.L.Lock()
	if this.destroyed.compareAndSwap(true, true) {
		this.cv.L.Unlock()
		return ErrAlreadyDestroyed
	}
	for atomic.LoadInt32(&this.numOperations) > 0 {
		this.cv.Wait()
	}
	defer this.cv.L.Unlock()
	for _, child := range this.children {
		// children can destroy themselves, ignore error
		child.Destroy()
	}
	if this.destroyCallback != nil {
		return this.destroyCallback()
	}
	return nil
}

func (this *destroyable) Do(f func() (interface{}, error)) (interface{}, error) {
	this.cv.L.Lock()
	if this.destroyed.value() {
		this.cv.L.Unlock()
		return nil, ErrAlreadyDestroyed
	}
	atomic.AddInt32(&this.numOperations, 1)
	this.cv.L.Unlock()
	value, err := f()
	atomic.AddInt32(&this.numOperations, -1)
	this.cv.Signal()
	return value, err
}

func (this *destroyable) AddChild(destroyable Destroyable) error {
	this.cv.L.Lock()
	if this.destroyed.value() {
		this.cv.L.Unlock()
		return ErrAlreadyDestroyed
	}
	defer this.cv.L.Unlock()
	this.children = append(this.children, destroyable)
	return nil
}
