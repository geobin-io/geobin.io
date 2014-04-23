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

	manageMap(t, mgr, func(mp map[string]interface{}) {
		mp["test_key"] = "test_val"
	})

	var val interface{}
	manageMap(t, mgr, func(mp map[string]interface{}) {
		val = mp["test_key"]
	})

	test.Expect(t, val, "test_val")
}

// mostly for race condition testing
func TestConcurrentAccess(t *testing.T) {
	obj := make(map[string]interface{})
	mgr := NewManager(obj)

	var w sync.WaitGroup
	w.Add(2)
	// lotsa reads
	go func(){
		for i := 0; i < 100000; i++ {
			manageMap(t, mgr, func(mp map[string]interface{}) {
				_ = mp["test_key"]
			})
		}
		w.Done()
	}()

	// lotsa writes
	go func() {
		for i := 0; i < 100000; i++ {
			manageMap(t, mgr, func(mp map[string]interface{}) {
				mp["test_key"] = i
			})
		}
		w.Done()
	}()
	w.Wait()

	var val interface{}
	manageMap(t, mgr, func(mp map[string]interface{}) {
			val = mp["test_key"]
	})

	test.Expect(t, val, 99999)
}

func manageMap(t *testing.T, mgr Manager, f func(myMap map[string]interface{})) {
	mgr.Touch(func(o interface{}) {
		if obj, ok := o.(map[string]interface{}); ok {
			f(obj)
		} else {
			t.Error("Managed object could not be asserted back to original type.")
		}
	})
}
