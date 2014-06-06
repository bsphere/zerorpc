package zerorpc

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
