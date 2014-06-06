package zerorpc

import (
	"errors"
	"log"
	"time"
)

const (
	// Heartbeat frequency,
	// default is 5 seconds
	HeartbeatFrequency = 5 * time.Second
)

type State int

const (
	Open = iota
	Closed
)

var (
	ErrClosedChannel = errors.New("zerorpc/channel closed channel")
	ErrLostRemote    = errors.New("zerorpc/channel lost remote")
)

// Channel representation
type Channel struct {
	Id            string
	state         State
	socket        *Socket
	socketInput   chan *Event
	channelOutput chan *Event
	channelErrors chan error
	lastHeartbeat time.Time
}

// Returns a pointer to a new channel,
// it adds the channel to the socket's array of channels
//
// it initiates sending of heartbeats event on the channel as long as
// the channel is open
func (s *Socket) NewChannel() *Channel {
	c := Channel{
		Id:            "",
		state:         Open,
		socket:        s,
		socketInput:   make(chan *Event),
		channelOutput: make(chan *Event),
		channelErrors: make(chan error),
		lastHeartbeat: time.Now(),
	}

	go c.listen()
	go c.handleHeartbeats()

	s.Channels = append(s.Channels, &c)

	log.Printf("ZeroRPC socket created new channel %s", c.Id)

	return &c
}

// Close the channel,
// set it's state to closed and removes it from the socket's array of channels
func (ch *Channel) Close() {
	ch.state = Closed

	ch.socket.RemoveChannel(ch)

	close(ch.socketInput)
	close(ch.channelOutput)
	close(ch.channelErrors)

	log.Printf("Channel %s closed", ch.Id)
}

// Sends an event on the channel,
// it sets the event response_to header to the channel's id
// and sends the event on the socket the channel belongs to
func (ch *Channel) SendEvent(e *Event) error {
	if ch.state == Closed {
		return ErrClosedChannel
	}

	if ch.Id != "" {
		e.Header["response_to"] = ch.Id
	} else {
		ch.Id = e.Header["message_id"].(string)

		go ch.sendHeartbeats()
	}

	log.Printf("Channel %s sending event %s", ch.Id, e.Header["message_id"].(string))

	return ch.socket.SendEvent(e)
}

// Sends heartbeat events on the channel as long as the channel is open,
// the heartbeats interval is defined in HeartbeatFrequency,
// default is 5 seconds
func (ch *Channel) sendHeartbeats() {
	for {
		time.Sleep(HeartbeatFrequency)

		if ch.state == Closed {
			return
		}

		ev, err := NewHeartbeat()
		if err != nil {
			log.Printf(err.Error())

			return
		}

		//log.Printf("Channel %s prepared heartbeat event %s", ch.Id, ev.Header["message_id"].(string))

		if err := ch.SendEvent(ev); err != nil {
			log.Printf(err.Error())

			return
		}

		log.Printf("Channel %s sent heartbeat", ch.Id)
	}
}

func (ch *Channel) listen() {
	for {
		ev := <-ch.socketInput

		switch ev.Name {
		case "OK":
			ch.channelOutput <- ev
		case "ERR":
			ch.channelOutput <- ev
		case "_zpc_hb":
			log.Printf("Channel %s received heartbeat", ch.Id)
			ch.lastHeartbeat = time.Now()
		}
	}
}

func (ch *Channel) handleHeartbeats() {
	for {
		if time.Since(ch.lastHeartbeat) > 2*HeartbeatFrequency {
			ch.channelErrors <- ErrLostRemote
		}
	}
}
