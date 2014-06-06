package zerorpc

import (
	"errors"
	"log"
	"time"
)

const (
	// ZeroRPC timeout,
	// default is 30 seconds
	ZeroRPCTimeout = 30 * time.Second
)

var (
	ErrZeroRPCTimeout = errors.New("zerorpc/client timeout")
)

// ZeroRPC client representation,
// it holds a pointer to the ZeroMQ socket
type Client struct {
	socket *Socket
}

// Connects to a ZeroRPC endpoint and returns a pointer to the new client
func NewClient(endpoint string) (*Client, error) {
	s, err := Connect(endpoint)
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
	return c.socket.Close()
}

func timeoutCounter(d time.Duration, done chan bool) {
	time.Sleep(d)
	done <- true
}

// Invokes a ZeroRPC method,
// name is the method name,
// args are the method arguments
//
// it returns the ZeroRPC response event on success
//
// if the ZeroRPC server raised an exception,
// it's name is returned as the err string along with the response event,
// the additional exception text and traceback can be found in the response event args
//
// it returns ErrZeroRPCTimeout if the ZeroRPC response timeouts,
// the timeout duration is defined in ZeroRPCTimeout, default is 30 seconds
func (c *Client) Invoke(name string, args ...interface{}) (*Event, error) {
	log.Printf("ZeroRPC client invoked %s with args %s", name, args)

	ev, err := NewEvent(name, args)
	if err != nil {
		return nil, err
	}

	ch := c.socket.NewChannel()
	defer ch.Close()

	err = ch.SendEvent(ev)
	if err != nil {
		return nil, err
	}

	timeout := make(chan bool)
	go timeoutCounter(ZeroRPCTimeout, timeout)

	for {
		select {
		case response := <-ch.ch:
			if response.Name == "OK" {
				return response, nil
			} else if response.Name == "ERR" {
				return response, errors.New(response.Args[0].(string))
			} else {
				return nil, errors.New("zerorpc/client invalid response event name")
			}

		case <-timeout:
			return nil, ErrZeroRPCTimeout
		}
	}
}
