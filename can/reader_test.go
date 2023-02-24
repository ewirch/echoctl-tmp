package can_test

import (
	"echoctl/can"
	"github.com/go-daq/canbus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestReaderBlockingBehavior(t *testing.T) {
	t.Run("Reader exits on kill", func(t *testing.T) {
		t.Parallel()

		reader, _, _ := NewSocket()

		runAndKillReader(t, reader, func() {})
	})
}

func TestReaderFunction(t *testing.T) {
	t.Run("Receives frame and sends it to Dispatcher", func(t *testing.T) {
		t.Parallel()

		reader, socket, toDispatcher := NewSocket()

		runAndKillReader(t, reader, func() {
			frame := NewFrame()
			socket.Inbound() <- frame
			select {
			case sentFrame := <-toDispatcher:
				assert.Equal(t, frame, sentFrame, "Read and sent frames are not equal")
			case <-time.After(time.Second):
				assert.Fail(t, "Reader did not send frame in 1s")
			}
		})
	})
}

func NewFrame() canbus.Frame {
	frame := canbus.Frame{
		ID:   350,
		Data: []byte{0, 1, 2, 3, 4, 5},
		Kind: canbus.SFF,
	}
	return frame
}

func NewSocket() (can.Reader, SocketMock, <-chan canbus.Frame) {
	socket := NewSocketMock()
	toDispatcher := make(chan canbus.Frame, 1)
	reader := can.NewReader(socket, toDispatcher, zap.NewNop())
	return reader, socket, toDispatcher
}

func runAndKillReader(t *testing.T, reader can.Reader, f func()) {
	tmb := reader.Read()

	f()

	tmb.Kill(nil)
	select {
	case <-tmb.Dead():
	// success
	case <-time.After(time.Second):
		assert.Fail(t, "Reader did not exit in 1s")
	}
}
