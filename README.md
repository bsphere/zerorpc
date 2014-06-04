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
	zmq "github.com/pebbe/zmq4"
)

func main() {
	s, err := zmq.NewSocket(zmq.REQ)
	if err != nil {
		panic(err)
	}

	defer s.Close()

	err = s.Connect("tcp://0.0.0.0:4242")
	if err != nil {
		panic(err)
	}

	m, err := zerorpc.NewEvent("hello", "John")
	if err != nil {
		panic(err)
	}

	by, err := m.PackBytes()
	if err != nil {
		panic(err)
	}

	s.SendBytes(by, 0)

	by, err = s.RecvBytes(0)
	if err != nil {
		panic(err)
	}

	m, err = zerorpc.UnPackBytes(by)
	if err != nil {
		panic(err)
	}

	fmt.Println(m)
}
```