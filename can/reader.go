package can

import (
	"context"
	"fmt"
	"github.com/go-daq/canbus"
	"go.uber.org/zap"
	"gopkg.in/tomb.v2"
)

type reader struct {
	toDispatcher chan<- canbus.Frame
	socket       Socket
	tomb         *tomb.Tomb
	log          *zap.Logger
}

// The Reader runs in the background, reading can-bus frames from socket and passing them to the Dispatcher.
type Reader interface {
	Read() *tomb.Tomb
}

var _ Reader = (*reader)(nil)

func NewReader(socket Socket, toDispatcher chan<- canbus.Frame, log *zap.Logger) Reader {
	return &reader{
		toDispatcher: toDispatcher,
		socket:       socket,
		tomb:         new(tomb.Tomb),
		log:          log,
	}
}

func (r *reader) Read() *tomb.Tomb {
	r.tomb.Go(r.read)
	return r.tomb
}

func (r *reader) read() error {
	ctx := r.tomb.Context(context.Background())
	for {
		responseFrame, err := r.socket.RecvCtx(ctx)
		if err != nil {
			if err == context.Canceled {
				return nil
			}
			return err
		}
		r.log.Debug("can: received", zap.Uint32("id", responseFrame.ID), zap.Stringer("kind", responseFrame.Kind), zap.String("b", fmt.Sprintf("% X", responseFrame.Data)))

		select {
		case r.toDispatcher <- responseFrame:
		case <-r.tomb.Dying():
			return tomb.ErrDying
		}
	}
}
