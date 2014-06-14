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



It supports streaming responses:

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

    response, err := c.InvokeStream("streaming_range", 10, 20, 2)
    if err != nil {
        fmt.Println(err.Error())
    }

    for _, r := range response {
        fmt.Println(r)
    }
}
```

It also supports first class exceptions, in case of an exception, 
the error returned from Invoke() or InvokeStream() is the exception name
and the args of the returned event are the exception description and traceback.

The client sends heartbeat events every 5 seconds, if twp heartbeat events are missed,
the remote is considered as lost and an ErrLostRemote is returned.


Server:

```go
package main

import (
	"errors"
	"fmt"
	"github.com/bsphere/zerorpc"
	"time"
)

func main() {
	s, err := zerorpc.NewServer("tcp://0.0.0.0:4242")
	if err != nil {
		panic(err)
	}

	defer s.Close()

	h := func(v []interface{}) (interface{}, error) {
		time.Sleep(10 * time.Second)
		return "Hello, " + v[0].(string), nil
	}

	s.RegisterTask("hello", &h)

	s.Listen()
}
```

It also supports first class exceptions, in case of the handler function returns an error,
the args of the event passed to the client is an array which is [err.Error(), nil, nil]