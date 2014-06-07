package zerorpc

import (
	"errors"
	"log"
	"time"
)

const (
	// ZeroRPC timeout,
	// default is 30 seconds
	ZeroRPCTimeout = 1 * time.Minute
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
// it returns ErrLostRemote if the channel misses 2 heartbeat events,
// default is 10 seconds
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

// Invokes a streaming ZeroRPC method,
// name is the method name,
// args are the method arguments
//
// it returns an array of ZeroRPC response events on success
//
// if the ZeroRPC server raised an exception,
// it's name is returned as the err string along with the response event,
// the additional exception text and traceback can be found in the response event args
//
// it returns ErrLostRemote if the channel misses 2 heartbeat events,
// default is 10 seconds
func (c *Client) InvokeStream(name string, args ...interface{}) ([]*Event, error) {
	log.Printf("ZeroRPC client invoked %s with args %s in streaming mode", name, args)

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

	out := make([]*Event, 0)

	for {
		select {
		case response := <-ch.channelOutput:
			if response.Name == "ERR" {
				return []*Event{response}, errors.New(response.Args[0].(string))
			} else if response.Name == "OK" {
				return []*Event{response}, nil
			} else if response.Name == "STREAM" {
				out = append(out, response)
			} else {
				return out, nil
			}

		case err := <-ch.channelErrors:
			return nil, err
		}
	}
}
