package zerorpc

// Returns a pointer to a new heartbeat event for a channel,
// it returns ErrClosedChannel if the channel is closed
func (ch *Channel) NewHeartbeat() (*Event, error) {
	if ch.state == Closed {
		return nil, ErrClosedChannel
	}

	ev, err := NewEvent("_zpc_hb", nil)
	if err != nil {
		return nil, err
	}

	return ev, nil
}
