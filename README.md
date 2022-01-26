# gogroupimports

Similar to goimport, but makes the grouping clearer.

The gogroupimports tool organizes the Go imports and rewrites them according to the following:

- First group: The standard library packages
- Next group: 3rd party library packages
- Last group: local packages

example:
```go
import (
	// The standard library packages
	"context"
	"fmt"
	"testing"

	// 3rd party library packages
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	// local packages
	"github.com/atotto/awesome"
)
```

## Installation

```
go install github.com/atotto/gofiximports@latest
go install golang.org/x/tools/cmd/goimports@latest
```

## Example of use

before:
```go
import (
	"testing"

	"context"
	"fmt"
	
	"github.com/atotto/awesome"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	
	"google.golang.org/protobuf/proto"
)
```

```
gogroupimports -local github.com/atotto/awesome -w foo.go
```

after:
```go
import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/atotto/awesome"
)
```

### vs goimports

The [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) tool does not delete empty line. Like this:

```
goimports -local github.com/atotto/awesome -w foo.go
```

```go
import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/atotto/awesome"
	
	"google.golang.org/protobuf/proto"
)
```

