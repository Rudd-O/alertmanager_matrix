# env

`env` is a small package for setting variables from environment variables.
It works similarly to [`flag`][flag], and can be used in conjunction.

**Note**: `env` requires Go 1.18 or later.

## Usage.

The basic usage is similar to `flag`, but with the `Value` type that `flag` hides exposed:

```go
package main

import (
	"fmt"

	"gitlab.com/slxh/go/env"
)

func main() {
	var i int

	env.AddVar(&i, "NUMBER", 0, "Some number.")
	env.Parse()

	fmt.Println(i)
}
```

It can also be mixed with `flag`:

```go
package main

import (
	"flag"
	"fmt"
 
	"gitlab.com/slxh/go/env"
)

func main() {
	var i int

	flag.IntVar(&i, "number", 0, "Some number.")

	// Parse both flags and environment variables.
	// The flag `number` above can be set using `NUMBER`.
	// This is also shown when calling `-h`, `-help` or setting `HELP`.
	// Note that flag overrides env.
	env.ParseWithFlags() 

	fmt.Println(i)
}
```

[flag]: https://pkg.go.dev/flag
