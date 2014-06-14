// Package zerorpc provides a client/server Golang library for the ZeroRPC protocol,
//
// for additional info see http://zerorpc.dotcloud.com
package zerorpc

import (
	zmq "github.com/pebbe/zmq4"
	"log"
)

// ZeroRPC socket representation
type socket struct {
	zmqSocket    *zmq.Socket
	Channels     []*channel
	server       *Server
	socketErrors chan error
}

// Connects to a ZeroMQ endpoint and returns a pointer to a new znq.DEALER socket,
// a listener for incoming messages is invoked on the new socket
func connect(endpoint string) (*socket, error) {
	zmqSocket, err := zmq.NewSocket(zmq.DEALER)
	if err != nil {
		return nil, err
	}

	s := socket{
		zmqSocket:    zmqSocket,
		Channels:     make([]*channel, 0),
		socketErrors: make(chan error),
	}

	if err := s.zmqSocket.Connect(endpoint); err != nil {
		return nil, err
	}

	log.Printf("ZeroRPC socket connected to %s", endpoint)

	go s.listen()

	return &s, nil
}

// Binds to a ZeroMQ endpoint and returns a pointer to a new znq.ROUTER socket,
// a listener for incoming messages is invoked on the new socket
func bind(endpoint string) (*socket, error) {
	zmqSocket, err := zmq.NewSocket(zmq.ROUTER)
	if err != nil {
		return nil, err
	}

	s := socket{
		zmqSocket: zmqSocket,
		Channels:  make([]*channel, 0),
	}

	if err := s.zmqSocket.Bind(endpoint); err != nil {
		return nil, err
	}

	log.Printf("ZeroRPC socket bound to %s", endpoint)

	go s.listen()

	return &s, nil
}

// Close the socket,
// it closes all the channels first
func (s *socket) close() error {
	for _, c := range s.Channels {
		c.close()
	}

	log.Printf("ZeroRPC socket closed")
	return s.zmqSocket.Close()
}

// Removes a channel from the socket's array of channels
func (s *socket) removeChannel(c *channel) {
	channels := make([]*channel, 0)

	for _, t := range s.Channels {
		if t != c {
			channels = append(channels, t)
		}
	}

	s.Channels = channels
}

// Sends an event on the ZeroMQ socket
func (s *socket) sendEvent(e *Event, identity string) error {
	b, err := e.PackBytes()
	if err != nil {
		return err
	}

	log.Printf("ZeroRPC socket sent event %s", e.Header["message_id"].(string))

	i, err := s.zmqSocket.SendMessage(identity, b)

	if err != nil {
		return err
	}

	log.Printf("ZeroRPC socket sent %d bytes", i)

	return nil
}

func (s *socket) listen() {
	log.Printf("ZeroRPC socket listening for incoming data")

	for {
		barr, err := s.zmqSocket.RecvMessageBytes(0)
		if err != nil {
			s.socketErrors <- err
		}

		t := 0
		for _, k := range barr {
			t += len(k)
		}

		log.Printf("ZeroRPC socket received %d bytes", t)

		ev, err := UnPackBytes(barr[len(barr)-1])
		if err != nil {
			s.socketErrors <- err
		}

		log.Printf("ZeroRPC socket recieved event %s", ev.Header["message_id"].(string))

		var ch *channel
		if _, ok := ev.Header["response_to"]; !ok {
			ch = s.newChannel(ev.Header["message_id"].(string))
			go ch.sendHeartbeats()

			if len(barr) > 1 {
				ch.identity = string(barr[0])
			}
		} else {
			for _, c := range s.Channels {
				if c.Id == ev.Header["response_to"].(string) {
					ch = c
				}
			}
		}

		if ch != nil && ch.state == open {
			log.Printf("ZeroRPC socket routing event %s to channel %s", ev.Header["message_id"].(string), ch.Id)

			ch.socketInput <- ev
		}
	}
}
