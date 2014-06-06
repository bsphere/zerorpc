package zerorpc

import (
	"errors"
	"time"
	"log"
)

const (
	// Heartbeat frequenct,
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
)

// Channel representation
type Channel struct {
	Id string
	state State
	socket *Socket
	ch chan *Event
}

func (ch *Channel) Close() {
	ch.state = Closed

	ch.socket.removeChannel(ch)

	log.Printf("Channel %s closed", ch.Id)
}

func (ch *Channel) SendEvent(e *Event) error {
	if ch.state == Closed {
		return ErrClosedChannel
	}

	if ch.Id != "" {
		e.Header["response_to"] = ch.Id
	} else {
		ch.Id = e.Header["message_id"].(string)
	}

	log.Printf("Channel %s sending event %s", ch.Id, e.Header["message_id"].(string))

	return ch.socket.SendEvent(e)
}

func (ch *Channel) sendHeartbeats() {
	for {
		time.Sleep(HeartbeatFrequency)

		if ch.state == Closed {
			return
		}

		ev, err := ch.NewHeartbeat()
		if err != nil {
			return
		}

		log.Printf("Channel %s prepared heartbeat event %s", ch.Id, ev.Header["message_id"].(string))

		if err := ch.SendEvent(ev); err != nil {
			panic(err)
		}

		log.Printf("Channel %s sent heartbeat", ch.Id)
	}
}

