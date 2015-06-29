package zerorpc

import (
	uuid "github.com/nu7hatch/gouuid"
	"github.com/ugorji/go/codec"
)

// ZeroRPC protocol version
const ProtocolVersion = 3

var (
	mh codec.MsgpackHandle
)

func init() {
	mh.RawToString = true

}

// Event representation

type EventHeader struct {
	Id         string `codec:"message_id,omitempty"`
	ResponseTo string `codec:"response_to,omitempty"`
	Version    int    `codec:"v"`
}

type Event struct {
	Header *EventHeader
	Name   string
	Args   codec.MsgpackSpecRpcMultiArgs
}

// Returns a pointer to a new event,
// a UUID V4 message_id is generated
func newEvent(name string, args ...interface{}) (*Event, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	header := &EventHeader{Id: id.String(), Version: ProtocolVersion}

	e := Event{
		Header: header,
		Name:   name,
		Args:   args,
	}

	return &e, nil
}

// Packs an event into MsgPack bytes
func (e *Event) packBytes() ([]byte, error) {
	data := make([]interface{}, 2)
	data[0] = e.Header
	data[1] = e.Name

	for _, a := range e.Args {
		data = append(data, a)
	}

	var buf []byte

	enc := codec.NewEncoderBytes(&buf, &mh)
	if err := enc.Encode(data); err != nil {
		return nil, err
	}

	return buf, nil
}

// Unpacks an event fom MsgPack bytes
func unPackBytes(b []byte) (*Event, error) {
	var e Event
	dec := codec.NewDecoderBytes(b, &mh)

	err := dec.Decode(&e)
	if err != nil {
		return nil, err
	}

	return &e, nil
}

// Returns a pointer to a new heartbeat event
func newHeartbeatEvent() (*Event, error) {
	ev, err := newEvent("_zpc_hb", nil)
	if err != nil {
		return nil, err
	}

	return ev, nil
}

// converts an interface{} to a type
func convertValue(v interface{}) interface{} {
	var out interface{}

	switch t := v.(type) {
	case []byte:
		out = string(t)

	case []interface{}:
		for i, x := range t {
			t[i] = convertValue(x)
		}

		out = t

	case map[interface{}]interface{}:
		for key, val := range v.(map[interface{}]interface{}) {
			t[key] = convertValue(val)
		}

		out = t

	default:
		out = t
	}

	return out
}
