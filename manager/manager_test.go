package manager

import (
	"github.com/geoloqi/geobin-go/test"
	"testing"
	"sync"
)

func TestNewManager(t *testing.T) {
	obj := make([]byte, 0, 5)
	mgr := NewManager(obj)

	test.Refute(t, mgr, nil)
	_, ok := mgr.(Manager)
	test.Expect(t, ok, true)
}

func TestBasicAccess(t *testing.T) {
	obj := make(map[string]interface{})
	mgr := NewManager(obj)

	mgr.Touch(func(o interface{}) {
		if obj, ok := o.(map[string]interface{}); ok {
			obj["test_key"] = "test_val"
		} else {
			t.Error("Managed object could not be asserted back to original type.")
		}
	})

	var val interface{}
	mgr.Touch(func(o interface{}) {
		if obj, ok := o.(map[string]interface{}); ok {
			val = obj["test_key"]
		} else {
			t.Error("Managed object could not be asserted back to original type.")
		}
	})

	test.Expect(t, val, "test_val")
}

// mostly for race condition testing
func TestConcurrentAccess(t *testing.T) {
	obj := make(map[string]interface{})
	mgr := NewManager(obj)

	var w sync.WaitGroup
	w.Add(2)
	// lotsa writes
	go func() {
		for i := 0; i < 100; i++ {
			mgr.Touch(func(o interface{}) {
				if obj, ok := o.(map[string]interface{}); ok {
					obj["test_key"] = i
				} else {
					t.Error("Managed object could not be asserted back to original type.")
				}
			})
		}
		w.Done()
	}()

	// lotsa reads
	go func(){
		for i := 0; i < 100; i++ {
			mgr.Touch(func(o interface{}) {
				if obj, ok := o.(map[string]interface{}); ok {
					_ = obj["test_key"]
				} else {
					t.Error("Managed object could not be asserted back to original type.")
				}
			})
		}
		w.Done()
	}()
	w.Wait()

	var val interface{}
	mgr.Touch(func(o interface{}) {
		if obj, ok := o.(map[string]interface{}); ok {
			val = obj["test_key"]
		} else {
			t.Error("Managed object could not be asserted back to original type.")
		}
	})

	test.Expect(t, val, 99)
}
