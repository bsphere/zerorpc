zerorpc
=======

Golang ZeroRPC client/server library - http://zerorpc.dotcloud.com

[![GoDoc](https://godoc.org/github.com/bsphere/zerorpc?status.png)](https://godoc.org/github.com/bsphere/zerorpc)


THE LIBRARY IS IN A VERY EARLY STAGE!!!


Usage
-----
Installation: `go get github.com/bsphere/zerorpc`

```go
package main

import (
	"fmt"
	"github.com/bsphere/zerorpc"
)

func main() {
	c, err := zerorpc.NewClient("tcp://0.0.0.0:4242")
    if err != nil {
        panic(err)
    }
    
    defer c.Close()
    
    response, err := c.Invoke("hello", "John")
    if err != nil {
        panic(err)
    }
    	
    fmt.Println(response)
}
```