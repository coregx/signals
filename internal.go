package signals

import "reflect"

// trackDependencyHelper is a shared helper for subscribing to dependencies with type erasure.
// It handles the complexity of subscribing to ReadonlySignal[X] where X is unknown at compile time.
//
// This helper is used by both Computed and Effect to avoid code duplication.
// The onChange callback is called whenever the dependency changes.
func trackDependencyHelper(dep any, onChange func()) Unsubscribe {
	// Helper interface to extract SubscribeForever
	type subscriber interface {
		SubscribeForever(fn func(any)) Unsubscribe
	}

	// First, try direct interface assertion
	if sub, ok := dep.(subscriber); ok {
		return sub.SubscribeForever(func(_ any) {
			onChange()
		})
	}

	// If that fails, handle common concrete types using type switch
	// This avoids reflection overhead for the most common types
	switch d := dep.(type) {
	case ReadonlySignal[int]:
		return d.SubscribeForever(func(_ int) { onChange() })
	case ReadonlySignal[string]:
		return d.SubscribeForever(func(_ string) { onChange() })
	case ReadonlySignal[bool]:
		return d.SubscribeForever(func(_ bool) { onChange() })
	case ReadonlySignal[float64]:
		return d.SubscribeForever(func(_ float64) { onChange() })
	case ReadonlySignal[int64]:
		return d.SubscribeForever(func(_ int64) { onChange() })
	default:
		// For any other type, use reflection as a fallback
		return subscribeAnyType(dep, onChange)
	}
}

// subscribeAnyType uses reflection to subscribe to any ReadonlySignal[X] type.
// This is a fallback for types not covered by the type switch in trackDependencyHelper.
func subscribeAnyType(dep any, onChange func()) Unsubscribe {
	// Use reflection to call SubscribeForever on any ReadonlySignal type
	val := reflect.ValueOf(dep)
	if !val.IsValid() {
		return func() {}
	}

	// Look for SubscribeForever method
	method := val.MethodByName("SubscribeForever")
	if !method.IsValid() {
		return func() {}
	}

	// Validate method signature
	fnType := method.Type()
	if fnType.NumIn() != 1 || fnType.NumOut() != 1 {
		return func() {}
	}

	// Create a callback using reflection
	callbackType := fnType.In(0)
	callback := reflect.MakeFunc(callbackType, func(_ []reflect.Value) []reflect.Value {
		onChange()
		return nil
	})

	// Call SubscribeForever(callback)
	results := method.Call([]reflect.Value{callback})
	if len(results) != 1 {
		return func() {}
	}

	// Extract Unsubscribe function
	unsubVal := results[0]
	if unsub, ok := unsubVal.Interface().(Unsubscribe); ok {
		return unsub
	}

	return func() {}
}
