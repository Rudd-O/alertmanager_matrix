package env

import (
	"fmt"
	"strconv"
	"time"
)

// Constraint contains the constraint for the Add, AddVar, NewValue and NewValueVar functions.
// Other types are not supported by these functions.
type Constraint interface {
	bool |
		float32 | float64 |
		int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 |
		string |
		time.Duration | time.Time
}

// parse a string to a type constrained by Constraint.
//
//nolint:funlen,cyclop
func parse[T Constraint](str string) (r T, err error) { //nolint:ireturn
	switch v := (interface{})(&r).(type) {
	case *bool:
		*v, err = strconv.ParseBool(str)
	case *float32:
		var f float64
		f, err = strconv.ParseFloat(str, 64)
		*v = float32(f)
	case *float64:
		*v, err = strconv.ParseFloat(str, 64)
	case *int:
		var i int64
		i, err = strconv.ParseInt(str, 10, 64)
		*v = int(i)
	case *int8:
		var i int64
		i, err = strconv.ParseInt(str, 10, 8)
		*v = int8(i)
	case *int16:
		var i int64
		i, err = strconv.ParseInt(str, 10, 16)
		*v = int16(i)
	case *int32:
		var i int64
		i, err = strconv.ParseInt(str, 10, 32)
		*v = int32(i)
	case *int64:
		*v, err = strconv.ParseInt(str, 10, 64)
	case *uint:
		var i uint64
		i, err = strconv.ParseUint(str, 10, 64)
		*v = uint(i)
	case *uint8:
		var i uint64
		i, err = strconv.ParseUint(str, 10, 8)
		*v = uint8(i)
	case *uint16:
		var i uint64
		i, err = strconv.ParseUint(str, 10, 16)
		*v = uint16(i)
	case *uint32:
		var i uint64
		i, err = strconv.ParseUint(str, 10, 32)
		*v = uint32(i)
	case *uint64:
		*v, err = strconv.ParseUint(str, 10, 64)
	case *string:
		*v = str
	case *time.Duration:
		*v, err = time.ParseDuration(str)
	case *time.Time:
		*v, err = time.Parse(time.RFC3339Nano, str)
	}

	if err != nil {
		return r, fmt.Errorf("cannot parse %T: %w", r, err)
	}

	return r, nil
}
