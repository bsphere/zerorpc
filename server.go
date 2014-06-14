package zerorpc

import (
	"errors"
	"log"
)

// ZeroRPC server representation,
// it holds a pointer to the ZeroMQ socket
type Server struct {
	socket   *Socket
	handlers []*TaskHandler
}

// Task handler representation
type TaskHandler struct {
	TaskName    string
	TaskHandler *func(args []interface{}) (interface{}, error)
}

var (
	ErrDuplicateHandler = errors.New("zerorpc/server duplicate task handler")
	ErrNoTaskHandler    = errors.New("zerorpc/server no handler for task")
)

/*
Binds to a ZeroRPC endpoint and returns a pointer to the new server

Usage example:

	package main

	import (
		"errors"
		"fmt"
		"github.com/bsphere/zerorpc"
		"time"
	)

	func main() {
		s, err := zerorpc.NewServer("tcp://0.0.0.0:4242")
		if err != nil {
			panic(err)
		}

		defer s.Close()

		h := func(v []interface{}) (interface{}, error) {
			time.Sleep(10 * time.Second)
			return "Hello, " + v[0].(string), nil
		}

		s.RegisterTask("hello", &h)

		s.Listen()
	}

It also supports first class exceptions, in case of the handler function returns an error,
the args of the event passed to the client is an array which is [err.Error(), nil, nil]
*/
func NewServer(endpoint string) (*Server, error) {
	s, err := Bind(endpoint)
	if err != nil {
		return nil, err
	}

	server := Server{
		socket:   s,
		handlers: make([]*TaskHandler, 0),
	}

	server.socket.server = &server

	return &server, nil
}

// Closes the ZeroMQ socket
func (s *Server) Close() error {
	return s.socket.Close()
}

// Register a task handler,
// it returns ErrDuplicateHandler if an handler was already registered for the task
func (s *Server) RegisterTask(name string, handlerFunc *func(args []interface{}) (interface{}, error)) error {
	for _, h := range s.handlers {
		if h.TaskName == name {
			return ErrDuplicateHandler
		}
	}

	s.handlers = append(s.handlers, &TaskHandler{TaskName: name, TaskHandler: handlerFunc})

	log.Printf("ZeroRPC server registered handler for task %s", name)

	return nil
}

// Invoke the handler for a task event,
// it returns ErrNoTaskHandler if no handler is registered for the task
func (s *Server) HandleTask(ev *Event) (interface{}, error) {
	for _, h := range s.handlers {
		if h.TaskName == ev.Name {
			log.Printf("ZeroRPC server handling task %s with args %s", ev.Name, ev.Args)

			return (*h.TaskHandler)(ev.Args)
		}
	}

	return nil, ErrNoTaskHandler
}

// Listen for incoming requests,
// it is a blocking function
func (s *Server) Listen() {
	for {
		err := <-s.socket.socketErrors
		if err != nil {
			log.Printf("ZeroRPC server error %s", err.Error())
		}
	}
}
