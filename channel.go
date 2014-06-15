package zerorpc

import (
	"errors"
	"log"
	"sync"
	"time"
)

const (
	// Heartbeat frequency,
	// default is 5 seconds
	HeartbeatFrequency = 5 * time.Second

	// Channel local buffer size,
	// default is 100
	BufferSize = 100
)

type state int

const (
	open = iota
	closed
)

var (
	errClosedChannel = errors.New("zerorpc/channel closed channel")
	ErrLostRemote    = errors.New("zerorpc/channel lost remote")
)

// Channel representation
type channel struct {
	Id            string
	state         state
	socket        *socket
	socketInput   chan *Event
	channelOutput chan *Event
	channelErrors chan error
	lastHeartbeat time.Time
	identity      string
	mu            sync.Mutex
}

// Returns a pointer to a new channel,
// it adds the channel to the socket's array of channels
//
// it initiates sending of heartbeats event on the channel as long as
// the channel is open
func (s *socket) newChannel(id string) *channel {
	s.mu.Lock()
	defer s.mu.Unlock()

	c := channel{
		Id:            id,
		state:         open,
		socket:        s,
		socketInput:   make(chan *Event, BufferSize),
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
func (ch *channel) close() {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	ch.state = closed

	ch.socket.removeChannel(ch)

	close(ch.socketInput)
	close(ch.channelOutput)
	close(ch.channelErrors)

	log.Printf("Channel %s closed", ch.Id)
}

// Sends an event on the channel,
// it sets the event response_to header to the channel's id
// and sends the event on the socket the channel belongs to
func (ch *channel) sendEvent(e *Event) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.state == closed {
		return errClosedChannel
	}

	if ch.Id != "" {
		e.Header["response_to"] = ch.Id
	} else {
		ch.Id = e.Header["message_id"].(string)

		go ch.sendHeartbeats()
	}

	log.Printf("Channel %s sending event %s", ch.Id, e.Header["message_id"].(string))

	identity := ch.identity

	return ch.socket.sendEvent(e, identity)
}

// Returns a pointer to a new "buffer size" event,
// it also returns the available space on the channel input buffer
func (ch *channel) newBufferSizeEvent() (*Event, int, error) {
	availSpace := ch.getFreeBufferSize()

	ev, err := newEvent("_zpc_more", []int{availSpace})
	if err != nil {
		return nil, 0, err
	}

	return ev, availSpace, nil
}

// Sends heartbeat events on the channel as long as the channel is open,
// the heartbeats interval is defined in HeartbeatFrequency,
// default is 5 seconds
func (ch *channel) sendHeartbeats() {
	for {
		time.Sleep(HeartbeatFrequency)

		if ch.state == closed {
			return
		}

		ev, err := newHeartbeatEvent()
		if err != nil {
			log.Printf(err.Error())

			return
		}

		if err := ch.sendEvent(ev); err != nil {
			log.Printf(err.Error())

			return
		}

		log.Printf("Channel %s sent heartbeat", ch.Id)
	}
}

func (ch *channel) listen() {
	streamCounter := 0

	for {
		if ch.state == closed {
			return
		}

		ev := <-ch.socketInput

		if ev == nil {
			continue
		}

		switch ev.Name {
		case "OK":
			ch.channelOutput <- ev

		case "ERR":
			ch.channelOutput <- ev

		case "STREAM":
			ch.channelOutput <- ev

			if streamCounter == 0 {
				me, free, err := ch.newBufferSizeEvent()
				if err == nil {
					streamCounter = free

					ch.sendEvent(me)

					streamCounter--
				}
			} else {
				streamCounter--
			}

		case "STREAM_DONE":
			ch.channelOutput <- ev

			streamCounter = 0

		case "_zpc_hb":
			log.Printf("Channel %s received heartbeat", ch.Id)
			ch.lastHeartbeat = time.Now()

		default:
			if ch.socket.server == nil {
				continue
			}

			log.Printf("Channel %s handling task %s with args %s", ch.Id, ev.Name, ev.Args)
			go func(ch *channel, ev *Event) {
				defer ch.close()
				defer ch.socket.removeChannel(ch)

				r, err := ch.socket.server.handleTask(ev)
				if err == nil {
					if e, err := newEvent("OK", []interface{}{r}); err != nil {
						log.Printf("zerorpc/channel %s", err.Error())
					} else {
						e := ch.sendEvent(e)
						if e != nil {
							ch.channelErrors <- e
						}
					}
				} else {
					if e, err2 := newEvent("ERR", []interface{}{err.Error(), nil, nil}); err2 == nil {
						e := ch.sendEvent(e)
						if e != nil {
							ch.channelErrors <- e
						}
					} else {
						log.Printf("zerorpc/channel %s", err.Error())
					}
				}
			}(ch, ev)
		}
	}
}

func (ch *channel) handleHeartbeats() {
	for {
		if ch.state == closed {
			return
		}

		if time.Since(ch.lastHeartbeat) > 2*HeartbeatFrequency {
			ch.channelErrors <- ErrLostRemote
		}

		time.Sleep(time.Second)
	}
}

func (ch *channel) getFreeBufferSize() int {
	return cap(ch.socketInput) - len(ch.socketInput)
}
