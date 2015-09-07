package zerorpc

import "testing"

func TestServerBind(t *testing.T) {
	s, err := NewServer("tcp://0.0.0.0:4242")
	if err != nil {
		panic(err)
	}

	defer s.Close()

	h := func(v []interface{}) (interface{}, error) {
		return "Hello, " + v[0].(string), nil
	}

	s.RegisterTask("hello", h)
}
