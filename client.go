package zerorpc

import (
	"errors"
	"log"
)

// ZeroRPC client representation,
// it holds a pointer to the ZeroMQ socket
type Client struct {
	socket *socket
}

// Connects to a ZeroRPC endpoint and returns a pointer to the new client
func NewClient(endpoint string) (*Client, error) {
	s, err := connect(endpoint)
	if err != nil {
		return nil, err
	}

	c := Client{
		socket: s,
	}

	return &c, nil
}

// Closes the ZeroMQ socket
func (c *Client) Close() error {
	return c.socket.close()
}

/*
Invokes a ZeroRPC method,
name is the method name,
args are the method arguments

it returns the ZeroRPC response event on success

if the ZeroRPC server raised an exception,
it's name is returned as the err string along with the response event,
the additional exception text and traceback can be found in the response event args

it returns ErrLostRemote if the channel misses 2 heartbeat events,
default is 10 seconds

Usage example:

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

It also supports first class exceptions, in case of an exception,
the error returned from Invoke() or InvokeStream() is the exception name
and the args of the returned event are the exception description and traceback.

The client sends heartbeat events every 5 seconds, if twp heartbeat events are missed,
the remote is considered as lost and an ErrLostRemote is returned.
*/
func (c *Client) Invoke(name string, args ...interface{}) (*Event, error) {
	log.Printf("ZeroRPC client invoked %s with args %s", name, args)

	ev, err := newEvent(name, args)
	if err != nil {
		return nil, err
	}

	ch := c.socket.newChannel("")
	defer ch.close()

	err = ch.sendEvent(ev)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case response := <-ch.channelOutput:
			if response.Name == "ERR" {
				return response, errors.New(response.Args[0].(string))
			} else {
				return response, nil
			}

		case err := <-ch.channelErrors:
			return nil, err
		}
	}
}

/*
Invokes a streaming ZeroRPC method,
name is the method name,
args are the method arguments

it returns an array of ZeroRPC response events on success

if the ZeroRPC server raised an exception,
it's name is returned as the err string along with the response event,
the additional exception text and traceback can be found in the response event args

it returns ErrLostRemote if the channel misses 2 heartbeat events,
default is 10 seconds

Usage example:

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

It also supports first class exceptions, in case of an exception,
the error returned from Invoke() or InvokeStream() is the exception name
and the args of the returned event are the exception description and traceback.

The client sends heartbeat events every 5 seconds, if twp heartbeat events are missed,
the remote is considered as lost and an ErrLostRemote is returned.
*/
func (c *Client) InvokeStream(name string, args ...interface{}) (chan *Event, error) {
	log.Printf("ZeroRPC client invoked %s with args %s in streaming mode", name, args)

	ev, err := newEvent(name, args)
	if err != nil {
		return nil, err
	}

	ch := c.socket.newChannel("")

	err = ch.sendEvent(ev)
	if err != nil {
		return nil, err
	}

	out := make(chan *Event)
	go func(out chan *Event, ch *channel) {
		defer close(out)
		defer ch.close()
		for {
			select {
			case response := <-ch.channelOutput:
				out <- response
				if response.Name != "STREAM" {
					return
				}
			case _ = <-ch.channelErrors:
				return
			}
		}
		return
	}(out, ch)
	return out, nil
}
