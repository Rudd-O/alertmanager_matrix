package env

import (
	"flag"
	"fmt"
	"reflect"
)

// Value is the interface to the value stored in an environment variable,
// or flag. It satisfies both flag.Getter and flag.Value, allowed it to be
// used with both.
type Value interface {
	// Get returns the current value.
	// It may return nil when there is no current value.
	Get() any

	// Default returns the default value as a string.
	// An empty string means there is no default value.
	Default() string

	// Set the value based on the given string.
	// An error is returned if the given string cannot be used to set the Value.
	Set(s string) error

	// String returns the string representation of the current value.
	// It may return an empty string when there is no current value.
	String() string
}

type value[T Constraint] struct {
	v *T
	d T
}

// NewValue creates a new Value with a default value, and returns
// a pointer to the value.
func NewValue[T Constraint](d T) (*T, Value) {
	v := d
	return &v, value[T]{v: &v, d: d}
}

// NewValueVar returns a Value for the given pointer and default value.
func NewValueVar[T Constraint](p *T, d T) Value {
	*p = d
	return &value[T]{v: p, d: d}
}

func (p *value[T]) ensureInit() {
	if p.v == nil {
		var v T
		p.v = &v
	}
}

func (p value[T]) Get() any {
	p.ensureInit()
	return *p.v
}

func (p value[T]) Default() string {
	return fmt.Sprintf("%v", p.d)
}

func (p value[T]) String() string {
	return fmt.Sprintf("%v", p.Get())
}

func (p value[T]) Set(s string) error {
	p.ensureInit()

	v, err := parse[T](s)
	if err != nil {
		return err
	}

	*p.v = v

	return err
}

// NewFuncValue returns a Value for the given function.
// Note that the Value will return nil for Get, and "" for Default and String.
func NewFuncValue(fn func(string) error) Value {
	return funcValue(fn)
}

type funcValue func(string) error

func (f funcValue) Get() any {
	return nil
}

func (f funcValue) Default() string {
	return ""
}

func (f funcValue) Set(s string) error {
	return f(s)
}

func (f funcValue) String() string {
	return ""
}

type getterValue struct {
	flag.Getter
	d string
}

func (g *getterValue) Default() string {
	return g.d
}

// ValueFromGetter converts a flag.Getter to a Value.
func ValueFromGetter(f flag.Getter, defValue string) Value {
	return &getterValue{Getter: f, d: defValue}
}

type flagValue struct {
	flag.Value
	d string
}

func (f *flagValue) Get() any {
	return nil
}

func (f *flagValue) Default() string {
	return f.d
}

// ValueFromFlag converts a flag.Value to a Value.
func ValueFromFlag(f flag.Value, defValue string) Value {
	return &flagValue{Value: f, d: defValue}
}

func isZero(v Value) bool {
	varType := reflect.TypeOf(v)
	if varType == nil {
		return true
	}

	var zero reflect.Value

	if varType.Kind() == reflect.Ptr {
		zero = reflect.New(varType.Elem())
	} else {
		zero = reflect.Zero(varType)
	}

	return v.Default() == zero.Interface().(Value).Default() //nolint:forcetypeassert
}
