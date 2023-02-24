package can_test

import (
	"echoctl/can"
	"echoctl/conf"
	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestPollerBlockingBehavior(t *testing.T) {

	t.Run("exits on kill", func(t *testing.T) {
		t.Parallel()

		subscriptions := []can.Subscription{
			{
				Command: NewCommand(123),
				Delay:   3 * time.Second,
			},
		}
		poller, _, _, _ := NewPoller(subscriptions)
		runAndKillPoller(t, poller, func() {})
	})
}

func TestPollerSchedule(t *testing.T) {
	t.Run("send command following schedule", func(t *testing.T) {
		t.Parallel()

		subscriptions := []can.Subscription{
			{
				Command: NewCommand(123),
				Delay:   3 * time.Second,
			},
		}
		poller, socket, inbound, clk := NewPoller(subscriptions)
		runAndKillPoller(t, poller, func() {

			makePollerIterate(inbound)
			expectNoSentCommand(t, socket)

			clk.Add(time.Second)
			makePollerIterate(inbound)
			expectNoSentCommand(t, socket)

			clk.Add(time.Second)
			makePollerIterate(inbound)
			expectNoSentCommand(t, socket)

			clk.Add(time.Second)
			expectCommand(t, socket, 123)
		})
	})

	t.Run("re-sends command following schedule", func(t *testing.T) {
		t.Parallel()

		subscriptions := []can.Subscription{
			{
				Command: NewCommand(123),
				Delay:   3 * time.Second,
			},
			{
				Command: NewCommand(456),
				Delay:   7 * time.Second,
			},
		}
		poller, socket, inbound, clk := NewPoller(subscriptions)
		runAndKillPoller(t, poller, func() {
			// synchronize go-routines: wait until poller gets in to the loop
			makePollerIterate(inbound)

			clk.Add(2 * time.Second)
			// 00:02
			makePollerIterate(inbound)
			expectNoSentCommand(t, socket)

			clk.Add(time.Second)
			// 00:03
			expectCommand(t, socket, 123)

			clk.Add(2 * time.Second)
			// 00:05
			makePollerIterate(inbound)
			expectNoSentCommand(t, socket)

			clk.Add(time.Second)
			// 00:06
			expectCommand(t, socket, 123)

			clk.Add(1 * time.Second)
			// 00:07
			expectCommand(t, socket, 456)

			clk.Add(1 * time.Second)
			// 00:08
			makePollerIterate(inbound)
			expectNoSentCommand(t, socket)

			clk.Add(1 * time.Second)
			// 00:09
			expectCommand(t, socket, 123)

			clk.Add(1 * time.Second)
			// 00:10
			makePollerIterate(inbound)
			expectNoSentCommand(t, socket)
		})
	})

	t.Run("sends commands in schedule order", func(t *testing.T) {
		t.Parallel()

		subscriptions := []can.Subscription{
			{
				Command: NewCommand(123),
				Delay:   2 * time.Second,
			},
			{
				Command: NewCommand(456),
				Delay:   3 * time.Second,
			},
		}
		poller, socket, inbound, clk := NewPoller(subscriptions)
		runAndKillPoller(t, poller, func() {
			// synchronize go-routines: wait until poller gets in to the loop
			makePollerIterate(inbound)

			clk.Add(2 * time.Second)
			expectCommand(t, socket, 123)

			clk.Add(time.Second)
			expectCommand(t, socket, 456)
		})
	})

	t.Run("inbound command changes schedule order", func(t *testing.T) {
		t.Parallel()

		subscriptions := []can.Subscription{
			{
				Command: NewCommand(123),
				Delay:   2 * time.Second,
			},
		}
		poller, socket, inbound, clk := NewPoller(subscriptions)
		runAndKillPoller(t, poller, func() {
			// synchronize go-routines: wait until poller gets in to the loop
			makePollerIterate(inbound)

			clk.Add(time.Second)
			inbound <- NewCommand(123)
			makePollerIterate(inbound)

			clk.Add(time.Second)
			expectNoSentCommand(t, socket)

			clk.Add(time.Second)
			expectCommand(t, socket, 123)
		})
	})
}

func expectNoSentCommand(t *testing.T, socket SocketMock) {
	select {
	case <-socket.Outbound():
		assert.Fail(t, "Command sent too early")
	default:
	}
}

func expectCommand(t *testing.T, socket SocketMock, canId conf.CanId) {
	select {
	case cmd := <-socket.Outbound():
		assert.Equal(t, uint32(canId), cmd.ID, "Looks like we received the wrong command")
	case <-time.After(time.Second):
		assert.Fail(t, "Poller failed to send command in 1s")
	}
}

// makePollerIterate() solves the problem of synchronising the test go routine and the Poller go routine. The Poller main loop consumes from the inbound channel and from the send timer channel. inbound is created as an unbuffered channel. The finished send operation guarantees that the main loop iterated.
func makePollerIterate(inbound chan<- conf.Command) {
	inbound <- NewUnknownCommand()
	inbound <- NewUnknownCommand()
}

func NewPoller(subscriptions []can.Subscription) (can.Poller, SocketMock, chan conf.Command, *clock.Mock) {
	socket := NewSocketMock()
	inbound := make(chan conf.Command)
	clck := clock.NewMock()
	poller := can.NewPoller(socket, subscriptions, inbound, clck, zap.NewNop())
	return poller, socket, inbound, clck
}

func runAndKillPoller(t *testing.T, poller can.Poller, f func()) {
	tmb := poller.Poll()
	f()
	tmb.Kill(nil)
	select {
	case <-tmb.Dead():
		// success
	case <-time.After(time.Second):
		assert.Fail(t, "Poller failed to exit in 1s")
	}
}

func NewCommand(canId conf.CanId) conf.Command {
	return conf.Command{
		Id: "001",
		Request: conf.RequestCommand{
			CanId:        canId,
			CommandBytes: []byte{3, 7, 5},
		},
	}
}
func NewUnknownCommand() conf.Command {
	return conf.Command{
		Id: "654654",
		Request: conf.RequestCommand{
			CanId:        654654,
			CommandBytes: []byte{3, 7, 5},
		},
	}
}
