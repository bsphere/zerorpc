package zerorpc

import (
	"errors"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/ugorji/go/codec"
)

// ZeroMQ protocol version
const ProtocolVersion = 3

// Event representation
type Event struct {
	Header map[string]interface{}
	Name   string
	Args   []interface{}
}

// Returns a pointer to a new event,
// a UUID V4 message_id is generated
func NewEvent(name string, args ...interface{}) (*Event, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	header := make(map[string]interface{})
	header["message_id"] = id.String()
	header["v"] = ProtocolVersion

	e := Event{
		Header: header,
		Name:   name,
		Args:   args,
	}

	return &e, nil
}

// Packs an event into MsgPack bytes
func (e *Event) PackBytes() ([]byte, error) {
	data := make([]interface{}, 2)
	data[0] = e.Header
	data[1] = e.Name
	//data[2] = e.Args
	for _, a := range e.Args {
		data = append(data, a)
	}

	var buf []byte

	enc := codec.NewEncoderBytes(&buf, &codec.MsgpackHandle{})
	if err := enc.Encode(data); err != nil {
		return nil, err
	}

	return buf, nil
}

// Unpacks an event fom MsgPack bytes
func UnPackBytes(b []byte) (*Event, error) {
	var mh codec.MsgpackHandle
	var v interface{}

	dec := codec.NewDecoderBytes(b, &mh)

	err := dec.Decode(&v)
	if err != nil {
		return nil, err
	}

	// get the event headers
	h, ok := v.([]interface{})[0].(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("zerorpc/event interface conversion error")
	}

	header := make(map[string]interface{})

	for k, v := range h {
		switch t := v.(type) {
		case []byte:
			header[k.(string)] = string(t)

		default:
			header[k.(string)] = t
		}
	}

	// get the event name
	n, ok := v.([]interface{})[1].([]byte)
	if !ok {
		return nil, errors.New("zerorpc/event interface conversion error")
	}

	// converts an interface{} to a type
	convertValue := func(v interface{}) interface{} {
		var out interface{}

		switch t := v.(type) {
		case []byte:
			out = string(t)

		default:
			out = t
		}

		return out
	}

	// get the event args
	args := make([]interface{}, 0)

	for i := 2; i < len(v.([]interface{})); i++ {
		t := v.([]interface{})[i]

		switch t.(type) {
		case []interface{}:
			for _, a := range t.([]interface{}) {
				args = append(args, convertValue(a))
			}

		default:
			args = append(args, convertValue(t))
		}
	}

	e := Event{
		Header: header,
		Name:   string(n),
		Args:   args,
	}

	return &e, nil
}

// Returns a pointer to a new heartbeat event
func NewHeartbeatEvent() (*Event, error) {
	ev, err := NewEvent("_zpc_hb", nil)
	if err != nil {
		return nil, err
	}

	return ev, nil
}
