package zerorpc

import "log"

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

	response := <-ch.ch

	return response, nil
}
