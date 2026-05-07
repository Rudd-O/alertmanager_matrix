package env

import (
	"cmp"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

var (
	// errUndefinedEnvironmentVariable is returned when an environment variable does not exist.
	errUndefinedEnvironmentVariable = errors.New("undefined environment variable")

	// errHelp is returned when HELP is set.
	errHelp = errors.New("env: help requested")
)

// ErrorHandling defines how EnvSet.Parse behaves when errors are encountered.
type ErrorHandling int

// Defined error handling flows.
const (
	ReturnLastError  ErrorHandling = iota // Continue parsing when errors are encountered (the default).
	ReturnFirstError                      // Return the first error that is encountered.
	ReturnAllErrors                       // Return all encountered errors using [errors.Join].
)

// EnvSet represents a set of defined environment variables.
type EnvSet struct { //nolint:revive // clashes with Set().
	prefix        string
	all           map[string]*Env
	set           map[string]*Env
	output        io.Writer
	parsed        bool
	errorHandling ErrorHandling
	helpFn        func()
}

// NewEnvSet returns an initialized EnvSet with the given prefix and ErrorHandling policy.
func NewEnvSet(prefix string, errorHandling ErrorHandling) *EnvSet {
	return &EnvSet{
		prefix:        prefix,
		errorHandling: errorHandling,
		all:           make(map[string]*Env),
		set:           make(map[string]*Env),
	}
}

// Add a definition for an environment variable of the specified name,
// with the provided UsageString string.
// Any value allowed by Constraint may be given.
// Note that only the suffix has to be provided as name if SetPrefix is used.
func (s *EnvSet) Add(name string, value Value, usage string) {
	s.Var(value, name, usage)
}

// FlagUsage returns a function that prints the usage of both
// the flags in the given [flag.FlagSet] and the current [EnvSet].
// It uses [flag.CommandLine] if the given FlagSet is nil.
func (s *EnvSet) FlagUsage(set *flag.FlagSet) func() {
	if set == nil {
		set = flag.CommandLine
	}

	return func() {
		if name := set.Name(); name != "" {
			_, _ = fmt.Fprintf(set.Output(), "Usage of %s:", os.Args[0])
		} else {
			_, _ = fmt.Fprint(set.Output(), "Usage:")
		}

		if haveFlags(set) {
			_, _ = fmt.Fprintln(set.Output(), "\nFlags:")
			set.PrintDefaults()
		}

		_, _ = fmt.Fprintln(set.Output(), "\nEnvironment variables:")
		_, _ = fmt.Fprintln(set.Output(), s.Usage())
	}
}

func haveFlags(set *flag.FlagSet) (h bool) {
	set.VisitAll(func(*flag.Flag) {
		h = true
	})

	return
}

// Func adds a definition for an environment variable of the specified name,
// with the provided UsageString string.
// The argument `fn` should point to a function that is called every time the
// named environment variable is encountered.
// Note that only the suffix has to be provided as name if SetPrefix is used.
func (s *EnvSet) Func(name, usage string, fn func(string) error) {
	s.Var(funcValue(fn), name, usage)
}

// Lookup returns the Env struct for an environment variable name.
// `nil` is returned if the variable is not defined.
func (s *EnvSet) Lookup(name string) *Env {
	if e, ok := s.all[name]; ok {
		return e
	}

	return nil
}

// Output returns the output that the EnvSet uses for methods like PrintDefaults.
func (s *EnvSet) Output() io.Writer {
	if s.output == nil {
		return os.Stderr
	}

	return s.output
}

// Parse parses the environment variables.
// An error is returned if environment variables contain invalid values.
func (s *EnvSet) Parse() error {
	errs := []error{nil}
	s.parsed = true

	if s.handleHelp() {
		return errHelp
	}

	for name, env := range s.all {
		if v, ok := os.LookupEnv(s.prefix + name); ok {
			s.set[name] = env

			if err := env.Value.Set(v); err != nil {
				err = fmt.Errorf("cannot set %s to %q: %w", name, v, err)

				switch s.errorHandling {
				case ReturnFirstError:
					return err
				case ReturnLastError:
					errs[0] = err
				case ReturnAllErrors:
					errs = append(errs, err)
				}
			}
		}
	}

	if len(errs) == 1 {
		return errs[0]
	}

	return errors.Join(errs...)
}

func (s *EnvSet) handleHelp() bool {
	if _, ok := os.LookupEnv(s.prefix + "HELP"); !ok || s.haveHelp() {
		return false
	}

	if s.helpFn != nil {
		s.helpFn()
	} else {
		s.printUsage()
	}

	return true
}

func (s *EnvSet) haveHelp() bool {
	_, ok := s.all["HELP"]

	return ok
}

func (s *EnvSet) printUsage() {
	_, _ = fmt.Fprintln(s.Output(), "Usage:")
	_, _ = fmt.Fprintln(s.Output(), s.Usage())
}

// Parsed returns a boolean indicating if Parse() has been called.
func (s *EnvSet) Parsed() bool {
	return s.parsed
}

// ParseWithFlagSet defines the flags in the given flag set as environment variables,
// and parses both the environment variables and flags.
// See SetFlag for details on the conversion between flags and environment variables.
func (s *EnvSet) ParseWithFlagSet(flagSet *flag.FlagSet, arguments []string) error {
	flagSet.VisitAll(s.SetFlag)

	if err := s.Parse(); err != nil {
		return fmt.Errorf("error parsing environment variables: %w", err)
	}

	if err := flagSet.Parse(arguments); err != nil {
		return fmt.Errorf("error parsing arguments: %w", err)
	}

	return nil
}

// SetFlag sets an environment variable based on a flag.Flag.
// The name of the flag is converted to an environment variable by converting
// the characters to uppercase, and replacing dashes (-) to underscores.
// For example: `this-flag` is converted to `THIS_FLAG`.
func (s *EnvSet) SetFlag(f *flag.Flag) {
	name := strings.ReplaceAll(strings.ToUpper(f.Name), "-", "_")

	var v Value
	if getter, ok := f.Value.(flag.Getter); ok {
		v = ValueFromGetter(getter, f.DefValue)
	} else {
		v = ValueFromFlag(f.Value, f.DefValue)
	}

	s.Var(v, name, f.Usage)
}

// SetFlagUsage set the usage on the given [flag.FlagSet] to [EnvSet.FlagUsage].
// It uses [flag.CommandLine] if the given FlagSet is nil.
func (s *EnvSet) SetFlagUsage(set *flag.FlagSet) {
	cmp.Or(set, flag.CommandLine).Usage = s.FlagUsage(set)
}

// PrintDefaults prints the UsageString string to the configured output (os.Stderr by default).
// Use SetOutput to change the output.
func (s *EnvSet) PrintDefaults() {
	_, _ = fmt.Fprintln(s.Output(), s.Usage())
}

// Set sets a variable to the given value.
func (s *EnvSet) Set(name, value string) error {
	env, ok := s.all[name]
	if !ok {
		return errUndefinedEnvironmentVariable
	}

	if err := env.Value.Set(value); err != nil {
		return fmt.Errorf("cannot set %s to %q: %w", name, value, err)
	}

	return nil
}

// SetOutput sets the output for functions that print, like PrintDefaults.
func (s *EnvSet) SetOutput(w io.Writer) {
	s.output = w
}

// Usage returns the usage string for the defined environment variables.
func (s *EnvSet) Usage() string {
	messages := make([]string, 0, len(s.all))

	for _, env := range s.all {
		messages = append(messages, env.UsageString(s.prefix))
	}

	sort.Strings(messages)

	return strings.Join(messages, "\n")
}

// Var adds a definition for an environment variable of the specified name,
// with the provided usage string.
// The type and value are contained in the first argument, of type Value.
// Note that only the suffix has to be provided as name if SetPrefix is used.
func (s *EnvSet) Var(v Value, name string, usage string) {
	s.all[name] = &Env{
		Value: v,
		Name:  name,
		Usage: usage,
	}
}

// Visit visits all the defined environment variables in lexicographical order.
// The given function is called for all variables that have been set.
func (s *EnvSet) Visit(fn func(*Env)) {
	for _, env := range sortEnvs(s.set) {
		fn(env)
	}
}

// VisitAll visits all the defined environment variables in lexicographical order.
// The given function is called for all variables, including ones not set.
func (s *EnvSet) VisitAll(fn func(*Env)) {
	for _, env := range sortEnvs(s.all) {
		fn(env)
	}
}

func sortEnvs(flags map[string]*Env) []*Env {
	list := make([]*Env, 0, len(flags))
	for _, env := range flags {
		list = append(list, env)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})

	return list
}
