package can

import (
	"context"
	"fmt"
	"github.com/go-daq/canbus"
	"golang.org/x/sys/unix"
)

// Socket is an abstraction around canbus.Socket, so it can be mocked in tests.
type Socket interface {
	Close() error
	Send(msg canbus.Frame) (int, error)
	RecvCtx(ctx context.Context) (msg canbus.Frame, err error)
}

func NewSocket(iface string) (Socket, error) {
	socket, err := canbus.New()
	if err != nil {
		return nil, fmt.Errorf("canbus.New(): %w", err)
	}
	err = socket.Bind(iface)
	if err != nil {
		return nil, fmt.Errorf("socket.Bind(\"%s\"): %w", iface, err)
	}
	err = socket.SetErrFilter(unix.CAN_ERR_MASK)
	if err != nil {
		return nil, fmt.Errorf("SetErrFilter(CAN_ERR_MASK): %w", err)
	}

	return socket, nil
}
