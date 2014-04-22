package test

import (
	"testing"
	"reflect"
)

// mustBeNil()/mustNotBeNil() isNillable() author: courtf
func MustBeNil(t *testing.T, a interface{}) {
	tp := reflect.TypeOf(a)

	if tp != nil && (!IsNillable(tp.Kind()) || !reflect.ValueOf(a).IsNil()) {
		t.Errorf("Expected %v (type %v) to be nil", a, tp)
	}
}

func MustNotBeNil(t *testing.T, a interface{}) {
	tp := reflect.TypeOf(a)

	if tp == nil || (IsNillable(tp.Kind()) && reflect.ValueOf(a).IsNil()) {
		t.Errorf("Expected %v (type %v) to not be nil", a, tp)
	}
}

func IsNillable(k reflect.Kind) (nillable bool) {
	kinds := []reflect.Kind{
		reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Ptr,
		reflect.Slice,
	}

	for i := 0; i < len(kinds); i++ {
		if kinds[i] == k {
			nillable = true
			break
		}
	}

	return
}

// Warning: directly comparing functions is unreliable
func Expect(t *testing.T, a interface{}, b interface{}) {
	btype := reflect.TypeOf(b)
	if b == nil {
		MustBeNil(t, a)
	} else if btype.Kind() == reflect.Func {
		if reflect.ValueOf(a).Pointer() != reflect.ValueOf(b).Pointer() {
			t.Errorf("Expected func %v (type %v) to equal func %v (type %v).", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
		}
	} else if !reflect.DeepEqual(a, b) {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

// Warning: directly comparing functions is unreliable
func Refute(t *testing.T, a interface{}, b interface{}) {
	btype := reflect.TypeOf(b)
	if b == nil {
		MustNotBeNil(t, a)
	} else if btype.Kind() == reflect.Func {
		if reflect.ValueOf(a).Pointer() == reflect.ValueOf(b).Pointer() {
			t.Errorf("Expected func %v (type %v) to not equal func %v (type %v).", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
		}
	} else if reflect.DeepEqual(a, b) {
		t.Errorf("Did not expect %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}
