# envi – Environment parser

## Installation

```sh
go get github.com/modernice/envi
```

## Usage

```go
package main

import (
	"log"

	"github.com/modernice/envi"
)

type Env struct {
	Foo string `env:"FOO"`
	Bar int    `env:"BAR"`
	Baz bool   `env:"BAZ"`
}

func main() {
	var env Env
	if err := envi.Parse(&env); err != nil {
		panic(err)
	}
}
```

## License

[MIT](LICENSE)
