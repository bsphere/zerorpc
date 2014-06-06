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
	data := make([]interface{}, 3)
	data[0] = e.Header
	data[1] = e.Name
	data[2] = e.Args

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

	// get the event headers
	a, ok := v.([]interface{})[2].([]interface{})
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

	args := make([]interface{}, 0)

	// handle both array of results and a single result
	for _, b := range a {
		x, ok := b.([]interface{})
		if ok {
			for _, v := range x {
				args = append(args, convertValue(v))
			}
		} else {
			args = append(args, convertValue(b))
		}
	}

	e := Event{
		Header: header,
		Name:   string(n),
		Args:   args,
	}

	return &e, nil
}
