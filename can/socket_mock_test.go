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
	NextSendError(error)
}
type socketMock struct {
	outbound      chan canbus.Frame
	inbound       chan canbus.Frame
	nextSendError *error
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
	if s.nextSendError == nil {
		s.outbound <- msg
		return 1, nil
	} else {
		err := *s.nextSendError
		s.nextSendError = nil
		return 0, err
	}
}

func (s *socketMock) RecvCtx(ctx context.Context) (msg canbus.Frame, err error) {
	select {
	case frame := <-s.inbound:
		return frame, nil
	case <-ctx.Done():
		return canbus.Frame{}, nil
	}
}

// NextSendError sets the error, which will be returned on next Send invocation.
func (s *socketMock) NextSendError(err error) {
	s.nextSendError = &err
}

func NewSocketMock() SocketMock {
	return &socketMock{
		outbound:      make(chan canbus.Frame, 1),
		inbound:       make(chan canbus.Frame, 1),
		nextSendError: nil,
	}
}
