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

type Client struct {
	socket *Socket
}

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

func (c *Client) Close() error {
	return c.socket.Close()
}

func timeoutCounter(d time.Duration, done chan bool) {
	time.Sleep(d)
	done <- true
}

func (c *Client) Invoke(name string, args ...interface{}) (*Event, error) {
	log.Printf("ZeroRPC client invoked %s with args %s", name, args)

	ev, err := NewEvent(name, args)
	if err != nil {
		return nil, err
	}

	ch := c.socket.newChannel()
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
			return response, nil

		case <-timeout:
			return nil, ErrZeroRPCTimeout
		}
	}
}
