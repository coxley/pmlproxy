# pmlproxy

This is a PlantUML gRPC server written in Go. Rendering is distributed to a set
of Java workers, using the `-pipe` option. As an add-on, groupcache can be
enabled to save JVM round-trips.

# Rationale

Will expand in this soon. Built as part of a larger project.

# Examples

The full interface is described in `pb/api.proto`.

**Basic server from within your own program**:

```go
package main

import (
    "fmt"
    "os"

    "github.com/coxley/pmlproxy/server"
)

func main() {
    // You're responsible for managing the pool, but enabling without one will
    // still cache on the local instance.
    server.DefaultHandler.GroupCache = true
    srv := &server.Server{Addr: "localhost:9000"}
    err := srv.ListenAndServe()
    if err != nil {
        fmt.Fprintf(os.Stderr, "caught error in server: %v", err)
    }
}
```

**Using the CLI**:

```bash
go get github.com/coxley/pmlproxy/pml
alias pml='pml --addr localhost:9000'

# Starting two servers with a shared cache
pml daemon --addr :8001 --cache-addr localhost:9001 -g localhost:9002
pml daemon --addr :8002 --cache-addr localhost:9002 -g localhost:9001

# Basic render
pml render diagram.pml > output.png
pml render -f SVG diagram.pml > output.svg

# Decode original text from image
pml extract output.png

# Diagram to short text
cat diagram.pml | pml shorten

# Misc

# Creates digaram-0.png and diagram-1.png
cat <<EOF | pml render --output-to-disk  # or -o
@startuml
rectangle Foo
rectangle Bar

Foo -> Bar: diagram-0
@enduml

@startuml
rectangle Foo
rectangle Bar

Baz -> Qux: diagram-1
@enduml
EOF
```

**Connecting to server from code**:

```go
package main

import (
    "fmt"
    "os"

	"google.golang.org/grpc"

	"github.com/coxley/pmlproxy/pb"
)

func getClient(addr string) (pb.PlantUMLClient, error) {
    // Don't actually use Insecure
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return pb.NewPlantUMLClient(conn), nil
}

var diagram = `
@startuml
rectangle Foo
@enduml
`

func main() {
    client, err := getClient(":9000")
    if err != nil {
        panic(err)
    }

    resp, err client.Render(&pb.RenderRequest{
        Diagram: &pb.Diagram{Source: diagram},
        Format: pb.Format_PNG,
    })
    if err != nil {
        panic(err)
    }
    // Multiple diagrams can be returned in one result, so index 0.
    // Ideally do a condition on length first.
    os.Stdout.Write(resp.Data[0])
}
```
