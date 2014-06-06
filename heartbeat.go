package zerorpc

// Returns a pointer to a new heartbeat event
func NewHeartbeat() (*Event, error) {
	ev, err := NewEvent("_zpc_hb", nil)
	if err != nil {
		return nil, err
	}

	return ev, nil
}
