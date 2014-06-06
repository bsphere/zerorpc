package zerorpc

import (
	zmq "github.com/pebbe/zmq4"
	"log"
	"bytes"
)

type Socket struct {
	zmqSocket *zmq.Socket
	Channels []*Channel
}

func Connect(endpoint string) (*Socket, error) {
	zmqSocket, err := zmq.NewSocket(zmq.DEALER)
	if err != nil {
		return nil, err
	}

	s := Socket{
		zmqSocket: zmqSocket,
		Channels: make([]*Channel, 0),
	}

	if err := s.zmqSocket.Connect(endpoint); err != nil {
		return nil, err
	}

	log.Printf("ZeroRPC socket connected to %s", endpoint)

	go s.listen()

	return &s, nil
}

func (s *Socket) Close() error {
	for _, c := range s.Channels {
		c.Close()
	}

	log.Printf("ZeroRPC socket closed")
	return s.zmqSocket.Close()
}

// Returns a pointer to a new channel
func (s *Socket) newChannel() *Channel {
	c := Channel{
		Id: "",
		state: Open,
		socket: s,
		ch: make(chan *Event),
	}

	s.Channels = append(s.Channels, &c)

	log.Printf("ZeroRPC socket created new channel %s", c.Id)

	//go c.sendHeartbeats()

	return &c
}

func (s *Socket) removeChannel(c *Channel) {
	channels := make([]*Channel, 0)

	for _, t := range s.Channels {
		if t != c {
			channels = append(channels, t)
		}
	}

	s.Channels = channels
}

func (s *Socket ) SendEvent(e *Event) error {
	b, err := e.PackBytes()
	if err != nil {
		return err
	}

	log.Printf("ZeroRPC socket sent event %s", e.Header["message_id"].(string))

	s.zmqSocket.Send("", zmq.SNDMORE)
	i, err := s.zmqSocket.SendMessage(b)
	if err != nil {
		return err
	}

	log.Printf("ZeroRPC socket send %d bytes", i)

	return nil
}

func (s *Socket) listen() {
	log.Printf("ZeroRPC socket listening for incoming data")

	for {
		s.zmqSocket.Recv(0)
		barr, err := s.zmqSocket.RecvMessageBytes(0)
		if err != nil {
			continue
		}

		b := bytes.NewBuffer(nil)

		for _, bt := range barr {
			if _, err := b.Write(bt); err != nil {
				continue
			}
		}

		log.Printf("ZeroRPC socket received %d bytes", len(b.Bytes()))

		ev, err := UnPackBytes(b.Bytes())
		if err != nil {
			continue
		}

		log.Printf("ZeroRPC socket recieved response event %s", ev.Header["message_id"].(string))

		for _, c := range s.Channels {
			if c.Id == ev.Header["response_to"].(string) {
				log.Printf("ZeroRPC socket routing event %s to channel %s", ev.Header["message_id"].(string),
					ev.Header["response_to"].(string))

				c.ch <- ev
			}
		}

	}
}
