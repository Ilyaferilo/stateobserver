package stateobserver

import (
	"context"
	"fmt"
	"reflect"
	"time"
)

// ObserverValue transmitted over the channel when the state of the object changes
type ObserverValue struct {
	Value interface{}
	Err   error
}

// ObserverValueI interface for internal using by StateObserver
// both methods executed only inside StateObserver goroutine
type ObserverValueI interface {
	// SetState invalidate cache and queued new state value into StateObserver goroutine
	SetState(newValue interface{}) error

	// GetState returns the remote state of some object
	GetState() ObserverValue
}

// The StateObserver used to synchronize the local state with the remote state.
type StateObserver struct {
	value ObserverValueI

	cache        ObserverValue
	stateChanged chan ObserverValue

	newState       chan interface{}
	stateTypeCheck func(interface{}) bool
}

// New returns the state observer for observer implementation
// observerInitValue is initial state for cache
func New(observerImpl ObserverValueI, observerInitValue interface{}) *StateObserver {

	observer := StateObserver{}
	observer.newState = make(chan interface{})
	observer.stateChanged = make(chan ObserverValue)
	observer.cache = ObserverValue{Value: observerInitValue, Err: nil}
	observer.value = observerImpl
	stateTypeStr := reflect.TypeOf(observerInitValue).String()
	observer.stateTypeCheck = func(i interface{}) bool {
		return reflect.TypeOf(i).String() == stateTypeStr
	}

	return &observer
}

// Notification occurs only if a new error has occurred
// for example, if there is no connection with the object, the notification will occur only at the first loss of communication
// precondition: e is not nil
func (o *StateObserver) checkError(e error) {
	var currentErr string
	if o.cache.Err != nil {
		currentErr = o.cache.Err.Error()
	}
	if currentErr != e.Error() {
		o.cache.Err = e
		o.cache.Value = []byte{}
		o.stateChanged <- o.cache
	}
}

func (o *StateObserver) tryRecover() {
	if r := recover(); r != nil {
		err := fmt.Errorf("observer was panicked: %v", r)
		o.checkError(err)
	}
}

func (o *StateObserver) setNewValue(val interface{}) {
	if reflect.DeepEqual(o.cache.Value, val) {
		return
	}
	defer o.tryRecover()
	o.cache = ObserverValue{val, nil}
	// notifying about the changed state
	o.stateChanged <- o.cache

	// Trying to set a new state
	if err := o.value.SetState(val); err != nil {
		o.checkError(err)
	}
}

// TODO use context for getState...
func (o *StateObserver) process(ctx context.Context) {

	defer o.tryRecover()
	current := o.value.GetState()
	if current.Err != nil {
		o.checkError(current.Err)
	} else if !reflect.DeepEqual(current.Value, o.cache.Value) {
		o.cache = current
		o.stateChanged <- o.cache
	}
}

// Start monitoring state of remote object.
// Method non-blocking.
// Returns channel for notification state changing of remote object
func (o *StateObserver) Start(ctx context.Context, pollTimeout time.Duration) <-chan ObserverValue {

	go func() {
		defer close(o.stateChanged)
		for {
			select {
			case <-ctx.Done():
				return

			case <-time.After(pollTimeout):
				o.process(ctx)

			case state := <-o.newState:
				o.setNewValue(state)
			}
		}
	}()

	return o.stateChanged
}

// SetState sets new state for internal observer cache
func (o *StateObserver) SetState(newValue interface{}) error {
	if !o.stateTypeCheck(newValue) {
		return fmt.Errorf("incorrect interface for observer: need %T; got %T",
			o.cache.Value, newValue)
	}
	o.newState <- newValue
	return nil
}
