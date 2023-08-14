// Package deepcopy makes deep copies of things. A standard copy will copy the
// pointers: deep copy copies the values pointed to.  Unexported field
// values are not copied.
package deepcopy

import (
	"fmt"
	"reflect"
	"time"
	"unsafe"
)

const (
	// startDetectingCyclesAfter is used to check circular reference once the counter exceeds it.
	startDetectingCyclesAfter = 1000
	
	// maxReferenceChainLength is used to avoid fatal error stack overflow if the reference chain is too long.
	maxReferenceChainLength = 1500
)

// Interface for delegating copy process to type
type Interface interface {
	DeepCopy() interface{}
}

// Copy returns a deep copy value of the original value.
// All the values are expected to be deep copied except some cases:
// 1. all unexposed fields in a struct won't be deep copied
// 2. channels are shared when deep copy
// 3. functions are shared when deep copy, beware of variables captured by your function when you do deep copy.
// 4. Copy will return errors for circular references and too long reference chains.
func Copy(src interface{}) (interface{}, error) {
	if src == nil {
		return nil, nil
	}
	// Make the interface a reflect.Value
	original := reflect.ValueOf(src)
	
	// Make a copy of the same type as the original.
	cpy := reflect.New(original.Type()).Elem()
	
	// Recursively copy the original.
	err := copyRecursive(original, cpy, &callState{
		ptrLevel: 0,
		ptrSeen:  make(map[interface{}]struct{})})
	if err != nil {
		return nil, err
	}
	
	// Return the copy as an interface.
	return cpy.Interface(), nil
}

func copyRecursive(original, cpy reflect.Value, state *callState) error {
	state.ptrLevel++
	defer func() {
		state.ptrLevel--
	}()
	if int(state.ptrLevel) > maxReferenceChainLength {
		return fmt.Errorf("excessive reference chain happened via %s", original.Type().String())
	}
	// check for implement Interface
	if original.CanInterface() {
		if copier, ok := original.Interface().(Interface); ok {
			cpy.Set(reflect.ValueOf(copier.DeepCopy()))
			return nil
		}
	}
	
	// handle according to original's Kind
	switch original.Kind() {
	case reflect.Ptr:
		ptr := original.Interface()
		// the condition is to eliminate cost for common cases. when circular reference,
		// the ptrLevel increases extremely fast and then only a little memory is needed
		// to be paid for checking.
		if state.ptrLevel > uint(startDetectingCyclesAfter) {
			if _, ok := state.ptrSeen[ptr]; ok {
				return fmt.Errorf("encountered a circular reference via %s", original.Type().String())
			}
			state.ptrSeen[ptr] = struct{}{}
			defer delete(state.ptrSeen, ptr)
		}
		// Get the actual value being pointed to.
		originalValue := original.Elem()
		// if it isn't valid, return.
		if !originalValue.IsValid() {
			return nil
		}
		cpy.Set(reflect.New(originalValue.Type()))
		
		err := copyRecursive(originalValue, cpy.Elem(), state)
		if err != nil {
			return err
		}
	case reflect.Interface:
		// If this is a nil, don't do anything
		if original.IsNil() {
			return nil
		}
		// Get the value for the interface, not the pointer.
		originalValue := original.Elem()
		// Get the value by calling Elem().
		copyValue := reflect.New(originalValue.Type()).Elem()
		err := copyRecursive(originalValue, copyValue, state)
		if err != nil {
			return err
		}
		cpy.Set(copyValue)
	case reflect.Struct:
		t, ok := original.Interface().(time.Time)
		if ok {
			cpy.Set(reflect.ValueOf(t))
			return nil
		}
		// Go through each field of the struct and copy it.
		for i := 0; i < original.NumField(); i++ {
			// The Type's StructField for a given field is checked to see if StructField.PkgPath
			// is set to determine if the field is exported or not because CanSet() returns false
			// for settable fields.  I'm not sure why.  -mohae
			if original.Type().Field(i).PkgPath != "" {
				continue
			}
			err := copyRecursive(original.Field(i), cpy.Field(i), state)
			if err != nil {
				return err
			}
		}
	
	case reflect.Slice:
		if state.ptrLevel > uint(startDetectingCyclesAfter) {
			// > A uintptr is an integer, not a reference. Converting a pointer
			// > to a uintptr creates an integer value with no pointer semantics.
			// > Even if a uintptr holds the address of some object, the garbage
			// > collector will not update that uintptr's value if the object
			// > moves, nor will that uintptr keep the object from being reclaimed
			//
			// Use unsafe.Pointer instead of uintptr because the runtime may
			// change its value when object is moved.
			//
			// The length is stored to distinguish the slice has been seen before
			// correctly to avoid cases like right fold a slice.
			ptr := struct {
				ptr    unsafe.Pointer
				length int
			}{unsafe.Pointer(original.Pointer()), original.Len()}
			
			if _, ok := state.ptrSeen[ptr]; ok {
				return fmt.Errorf("encountered a circular reference via %s", original.Type().String())
			}
			state.ptrSeen[ptr] = struct{}{}
			defer delete(state.ptrSeen, ptr)
		}
		
		if original.IsNil() {
			return nil
		}
		// Make a new slice and copy each element.
		cpy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		for i := 0; i < original.Len(); i++ {
			err := copyRecursive(original.Index(i), cpy.Index(i), state)
			if err != nil {
				return err
			}
		}
	case reflect.Array:
		// since origin is an array, the capacity of array will be conserved
		cpy.Set(reflect.New(original.Type()).Elem())
		for i := 0; i < original.Len(); i++ {
			err := copyRecursive(original.Index(i), cpy.Index(i), state)
			if err != nil {
				return err
			}
		}
	case reflect.Map:
		ptr := unsafe.Pointer(original.Pointer())
		if state.ptrLevel > uint(startDetectingCyclesAfter) {
			if _, ok := state.ptrSeen[ptr]; ok {
				return fmt.Errorf("encountered a circular reference via %s", original.Type().String())
			}
			state.ptrSeen[ptr] = struct{}{}
			defer delete(state.ptrSeen, ptr)
		}
		
		if original.IsNil() {
			return nil
		}
		cpy.Set(reflect.MakeMap(original.Type()))
		for _, key := range original.MapKeys() {
			originalValue := original.MapIndex(key)
			copyValue := reflect.New(originalValue.Type()).Elem()
			err := copyRecursive(originalValue, copyValue, state)
			if err != nil {
				return err
			}
			copiedKey := reflect.New(key.Type()).Elem()
			err = copyRecursive(key, copiedKey, state)
			if err != nil {
				return err
			}
			
			cpy.SetMapIndex(copiedKey, copyValue)
		}
	default:
		cpy.Set(original)
	}
	return nil
}

type callState struct {
	ptrLevel uint
	ptrSeen  map[interface{}]struct{}
}
