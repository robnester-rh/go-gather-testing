# go-gather

WIP, experimental, sharp edges.

## Examples 

### Copy file to file

```
package main

import (
	"context"
	"fmt"

	"github.com/enterprise-contract/go-gather/gather"
)

func main() {
	ctx := context.Background()
	metadata, err := gather.Gather(ctx, "file:///tmp/foo.txt", "file:///tmp/bar.txt")

	if err != nil {
		fmt.Println("error:", err)
	}

	fmt.Println(metadata.Get())
}
```

Executing the above,

``` bash
$ go run main.go
```
produces the following output:

```
map[path:file:///tmp/baz/bar.txt sha:ef4e93945f5b3d481abe655d6ce3870132994c0bd5840e312d7ac97cde021050 size:71680 timestamp:2024-04-25 09:18:20.978669581 -0400 EDT]
```

### Copy directory to directory
```
package main

import (
	"context"
	"fmt"

	"github.com/enterprise-contract/go-gather/gather"
)

func main() {
	ctx := context.Background()
	metadata, err := gather.Gather(ctx, "file:///tmp/foo", "file:///tmp/bar")

	if err != nil {
		fmt.Println("error:", err)
	}

	fmt.Println(metadata.Get())
}
```

Executing the above,

``` bash
$ go run main.go
```
produces the following output:

```
map[path:/tmp/bar/ size:0 timestamp:2024-04-25 09:57:20.863229843 -0400 EDT]
```

### Clone git repo to local filesystem

```
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/enterprise-contract/go-gather/gather"
)

func main() {
	ctx := context.Background()
	destination := "/tmp/repo"
	source := "git::git@github.com:example/example.git"
	metadata, err := gather.Gather(ctx, sourcep, destination)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	fmt.Printf(metadata.Get())
}
```
Executing the above,

``` bash
$ go run main.go
```
produces the following output:

```
 map[size:1024 path:/path/to/file.txt timestamp:2022-01-01 12:00:00 +0000 UTC commits:[{689da11ffaef9d523615b3518cb1f2916a37ec42 {J Doe jdoe@example.com 2022-01-01 12:00:00 +0000 +0000} {J Doe jdoe@example.com 2022-01-01 12:00:00 +0000 +0000} Add new shiny feature [58b071e48f6e9e81ede4f284ee2c2aeeb06b3625] UTF-8 0xc0000d62c0}] path: size:0 timestamp:0001-01-01 00:00:00 +0000 UTC]
```