package zerorpc

import (
	"errors"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/ugorji/go/codec"
)

// ZeroRPC protocol version
const ProtocolVersion = 3

// Event representation
type Event struct {
	Header map[string]interface{}
	Name   string
	Args   []interface{}
}

// Returns a pointer to a new event,
// a UUID V4 message_id is generated
func newEvent(name string, args ...interface{}) (*Event, error) {
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
func (e *Event) packBytes() ([]byte, error) {
	data := make([]interface{}, 2)
	data[0] = e.Header
	data[1] = e.Name

	for _, a := range e.Args {
		data = append(data, a)
	}

	var buf []byte

	enc := codec.NewEncoderBytes(&buf, &codec.MsgpackHandle{WriteExt:true})
	if err := enc.Encode(data); err != nil {
		return nil, err
	}

	return buf, nil
}

// Unpacks an event fom MsgPack bytes
func unPackBytes(b []byte) (*Event, error) {
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
