package mailgen

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"
)

var (
	// Initialize random seed
	// #nosec G404 -- non-cryptographic randomness is sufficient for template/example mail generation.
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// builtinFuncs returns the built-in template functions
func builtinFuncs() map[string]interface{} {
	return map[string]interface{}{
		"pickRandom": pickRandom,
	}
}

// pickRandom randomly selects an item from a slice
func pickRandom(items interface{}) (interface{}, error) {
	if items == nil {
		return nil, fmt.Errorf("cannot pick from nil")
	}

	v := reflect.ValueOf(items)

	// Handle pointer dereference
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, fmt.Errorf("cannot pick from nil pointer")
		}
		v = v.Elem()
	}

	// Convert interface arrays to proper type
	if v.Kind() == reflect.Interface {
		v = reflect.ValueOf(v.Interface())
	}

	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil, fmt.Errorf("cannot pick random item from type %T, expected slice or array", items)
	}

	length := v.Len()
	if length == 0 {
		return nil, fmt.Errorf("cannot pick from empty slice")
	}

	// Pick a random index
	idx := rnd.Intn(length)
	item := v.Index(idx)

	// Handle interface{} elements
	if item.Kind() == reflect.Interface {
		return item.Interface(), nil
	}

	return item.Interface(), nil
}
