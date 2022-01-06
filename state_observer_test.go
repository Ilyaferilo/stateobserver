package stateobserver

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

type testObserver struct {
	val []byte
}

func (b *testObserver) SetState(pi interface{}) error {
	p := reflect.ValueOf(pi).Bytes()
	if len(p) == 0 {
		panic("it`s very bad slice")
	}
	if p[0] == 0 {
		return fmt.Errorf("zero in begin of p is incorrect")
	}
	b.val = p
	return nil
}

func (b *testObserver) GetState() ObserverValue {
	if b.val[len(b.val)-1] == 0 {
		return ObserverValue{b.val, fmt.Errorf("zero in end is incorrect")}

	}
	return ObserverValue{b.val, nil}
}
func TestStateObserver(t *testing.T) {

	newValue := []byte{44, 42, 3}

	testVal := testObserver{val: []byte{1, 2, 3, 4}}
	o := New(&testVal, testVal.val)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ch := o.Start(ctx, 100*time.Millisecond)
	err := o.SetState(1)
	if err == nil {
		t.Errorf("observer Set does not check input type")
	}

	go func() {
		for {
			v, opened := <-ch
			if !opened {
				return
			}
			fmt.Println("new val = ", v)
		}
	}()

	time.Sleep(200 * time.Millisecond)

	// // panic set
	o.SetState([]byte{})

	// for fail get
	o.SetState([]byte{1, 0})
	time.Sleep(600 * time.Millisecond)

	// ok set
	o.SetState(newValue)
	// fail set
	o.SetState([]byte{0})

	<-ctx.Done()
	if !bytes.Equal(testVal.val, newValue) {
		t.Errorf("set new value was not worked")
	}
}

type SomeStruct struct {
	valInt  int
	valStr  string
	valFlag bool
}
type SomeStructObserver struct {
	v SomeStruct
}

func (b *SomeStructObserver) SetState(pi interface{}) error {
	str, _ := pi.(SomeStruct)
	b.v = str
	return nil
}

func (b *SomeStructObserver) GetState() ObserverValue {
	return ObserverValue{b.v, nil}
}

func TestStateObserverStruct(t *testing.T) {
	testVal := SomeStructObserver{}
	o := New(&testVal, SomeStruct{})
	ctx, cancel := context.WithTimeout(context.Background(), 2600*time.Millisecond)
	defer cancel()

	etalon := SomeStruct{valInt: 12, valStr: "xx", valFlag: false}
	ch := o.Start(ctx, 100*time.Millisecond)

	var counter int32 = 0

	go func() {
		for {
			_, opened := <-ch
			if !opened {
				return
			}

			atomic.AddInt32(&counter, 1)
		}
	}()

	o.SetState(etalon)
	o.SetState(etalon)

	<-ctx.Done()
	load := atomic.LoadInt32(&counter)
	if load != 1 {
		t.Errorf("observer failed: updates %v count", load)
	}

}
