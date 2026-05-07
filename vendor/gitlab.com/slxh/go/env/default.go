// Package env implements environment variable parsing.
package env

import (
	"cmp"
	"errors"
	"flag"
	"fmt"
)

// DefaultSet contains the default environment variable set.
//
//nolint:gochecknoglobals
var DefaultSet = NewEnvSet("", ReturnLastError)

// Add a definition for an environment variable of the specified name,
// with the provided UsageString string.
// Any value allowed by Constraint may be given.
// Note that only the suffix has to be provided as name if SetPrefix is used.
func Add[T Constraint](name string, value T, usage string) *T {
	p, v := NewValue(value)

	DefaultSet.Add(name, v, usage)

	return p
}

// AddVar adds a definition for an environment variable of the specified name,
// with the provided UsageString string.
// The argument `p` is a pointer to a variable that stores the value of the
// environment variable.
// Any value allowed by Constraint may be given.
// Note that only the suffix has to be provided as name if SetPrefix is used.
func AddVar[T Constraint](p *T, name string, value T, usage string) {
	DefaultSet.Add(name, NewValueVar(p, value), usage)
}

// FlagUsage returns a function that prints the usage of both
// the flags in the given [flag.FlagSet] and the default [EnvSet].
// It uses [flag.CommandLine] if the given FlagSet is nil.
func FlagUsage(set *flag.FlagSet) func() {
	return DefaultSet.FlagUsage(set)
}

// Func adds a definition for an environment variable of the specified name,
// with the provided UsageString string.
// The argument `fn` should point to a function that is called every time the
// named environment variable is encountered.
// Note that only the suffix has to be provided as name if SetPrefix is used.
func Func(name, usage string, fn func(string) error) {
	DefaultSet.Func(name, usage, fn)
}

// Lookup returns the Env struct for an environment variable name.
// `nil` is returned if the variable is not defined.
func Lookup(name string) *Env {
	return DefaultSet.Lookup(name)
}

// Parse parses the environment variables.
// An error is returned if environment variables contain invalid values.
func Parse() error {
	return DefaultSet.Parse()
}

// Parsed returns a boolean indicating if Parse() has been called.
func Parsed() bool {
	return DefaultSet.Parsed()
}

// ParseWithFlags defines the flags as set in the flag package as environment variables
// and parses both the environment variables and flags.
// It overrides [flag.Usage] to contain both flag and env usage using [SetFlagUsage].
// See [EnvSet.SetFlag] for details on the conversion between flags and environment variables.
func ParseWithFlags() error {
	flag.VisitAll(DefaultSet.SetFlag)
	SetFlagUsage(nil)

	DefaultSet.helpFn = FlagUsage(nil)

	if err := Parse(); err != nil {
		if errors.Is(err, errHelp) {
			return err
		}

		return fmt.Errorf("parse env: %w", err)
	}

	flag.Parse()

	return nil
}

// PrintDefaults prints the UsageString string to os.Stderr.
func PrintDefaults() {
	DefaultSet.PrintDefaults()
}

// Set sets a variable to the given value.
func Set(name, value string) error {
	return DefaultSet.Set(name, value)
}

// SetPrefix sets the prefix for all environment variables.
// This allows environment variables to share a common prefix.
func SetPrefix(prefix string) {
	DefaultSet.prefix = prefix
}

// SetErrorHandling sets the error handling for Parse.
// The default policy is the ReturnFirstError policy.
func SetErrorHandling(errorHandling ErrorHandling) {
	DefaultSet.errorHandling = errorHandling
}

// SetFlagUsage set the usage on the given [flag.FlagSet] to [EnvSet.FlagUsage].
// It uses [flag.CommandLine] if the given FlagSet is nil.
func SetFlagUsage(set *flag.FlagSet) {
	cmp.Or(set, flag.CommandLine).Usage = FlagUsage(set)
}

// Usage returns the usage string for the defined environment variables.
func Usage() string {
	return DefaultSet.Usage()
}

// Var adds a definition for an environment variable of the specified name,
// with the provided usage string.
// The type and value are contained in the first argument, of type Value.
// Note that only the suffix has to be provided as name if SetPrefix is used.
func Var(v Value, name string, usage string) {
	DefaultSet.Var(v, name, usage)
}

// Visit visits all the defined environment variables in lexicographical order.
// The given function is called for all variables that have been set.
func Visit(fn func(*Env)) {
	DefaultSet.Visit(fn)
}

// VisitAll visits all the defined environment variables in lexicographical order.
// The given function is called for all variables, including ones not set.
func VisitAll(fn func(*Env)) {
	DefaultSet.VisitAll(fn)
}
