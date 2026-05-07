package env

import (
	"fmt"
	"strings"
	"time"
)

// Env represents an environment variable.
type Env struct {
	// Name is the name of the environment variable.
	Name string

	// Usage is the basic usage string for the environment variable.
	Usage string

	// Value is the underlying Value of the environment variable.
	Value Value
}

// UsageString returns the usage string for the environment variable,
// along with any default values.
func (e *Env) UsageString(prefix string) string {
	kind, usage := e.UnquoteUsage()
	msg := "  " + prefix + e.Name + " " + kind + "\n    \t" + usage

	if !isZero(e.Value) {
		var format string

		switch e.Value.(type) {
		case value[string], *value[string]:
			format = "%q"
		default:
			format = "%v"
		}

		msg += fmt.Sprintf(" (default "+format+")", e.Value.Default())
	}

	return msg
}

// UnquoteUsage extracts a backtick-quoted kind from the usage string of an environment variable
// and returns it and the un-quoted usage.
// Given "a `name` to show" it returns ("name", "a name to show").
// If there are no back quotes, the name is based on the type of the environment variable.
func (e *Env) UnquoteUsage() (kind, usage string) {
	usage = e.Usage

	i := strings.IndexRune(usage, '`')
	if i >= 0 {
		if j := strings.IndexRune(usage[i+1:], '`'); j >= 0 {
			// kind is part between backticks.
			kind = usage[i+1 : i+j+1]

			// usage is string without backticks.
			usage = usage[:i] + kind + usage[i+j+2:]

			return kind, usage
		}
	}

	if e.Value == nil {
		return "nil", usage
	}

	switch t := e.Value.Get().(type) {
	case time.Duration:
		kind = "duration"
	case time.Time:
		kind = "timestamp"
	default:
		kind = fmt.Sprintf("%T", t)
	}

	return
}
