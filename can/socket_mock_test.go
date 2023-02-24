package can_test

import (
	"context"
	"echoctl/can"
	"github.com/go-daq/canbus"
)

type SocketMock interface {
	can.Socket
	Outbound() <-chan canbus.Frame
	Inbound() chan<- canbus.Frame
}
type socketMock struct {
	outbound chan canbus.Frame
	inbound  chan canbus.Frame
}

func (s *socketMock) Outbound() <-chan canbus.Frame {
	return s.outbound
}

func (s *socketMock) Inbound() chan<- canbus.Frame {
	return s.inbound
}

func (s *socketMock) Close() error {
	close(s.inbound)
	close(s.outbound)
	return nil
}

func (s *socketMock) Send(msg canbus.Frame) (int, error) {
	s.outbound <- msg
	return 1, nil
}

func (s *socketMock) RecvCtx(ctx context.Context) (msg canbus.Frame, err error) {
	select {
	case frame := <-s.inbound:
		return frame, nil
	case <-ctx.Done():
		return canbus.Frame{}, nil
	}
}

func NewSocketMock() SocketMock {
	return &socketMock{
		outbound: make(chan canbus.Frame, 1),
		inbound:  make(chan canbus.Frame, 1),
	}
}
